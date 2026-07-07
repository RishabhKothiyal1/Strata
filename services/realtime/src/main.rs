use axum::{
	extract::{ws, State, WebSocketUpgrade},
	response::{IntoResponse, Response},
	routing::get,
	Json, Router,
};
use futures_util::{sink::SinkExt, stream::StreamExt};
use serde::{Deserialize, Serialize};
use std::{
	collections::{HashMap, HashSet},
	net::SocketAddr,
	sync::{
		atomic::{AtomicUsize, Ordering},
		Arc,
	},
};
use tokio::sync::{mpsc, RwLock};
use tracing::{debug, error, info};

static CLIENT_ID_COUNTER: AtomicUsize = AtomicUsize::new(1);

#[derive(Clone)]
struct AppState {
	// Map from Room Name -> HashMap of ClientID -> UnboundedSender
	registry: Arc<RwLock<HashMap<String, HashMap<usize, mpsc::UnboundedSender<String>>>>>,
	nats_client: async_nats::Client,
}

#[derive(Deserialize, Serialize, Debug)]
#[serde(tag = "event", rename_all = "lowercase")]
enum ClientEvent {
	Subscribe {
		channel: String,
	},
	Unsubscribe {
		channel: String,
	},
	Broadcast {
		channel: String,
		data: serde_json::Value,
	},
}

#[derive(Serialize)]
struct HealthResponse {
	status: String,
	service: String,
}

#[tokio::main]
async fn main() {
	// Initialize tracing logger
	tracing_subscriber::fmt()
		.with_env_filter(
			tracing_subscriber::EnvFilter::try_from_default_env()
				.unwrap_or_else(|_| "realtime=debug,axum=info".into()),
		)
		.init();

	info!("Starting Strata Realtime Service in Rust...");

	// Load Configuration
	let port = std::env::var("PORT").unwrap_or_else(|_| ":8083".to_string());
	let nats_url = std::env::var("NATS_URL").unwrap_or_else(|_| "nats://nats:4222".to_string());

	// Connect to NATS message broker
	let nats_client = async_nats::connect(&nats_url)
		.await
		.expect("Failed to connect to NATS message broker");
	info!("Successfully connected to NATS at {}", nats_url);

	let registry = Arc::new(RwLock::new(HashMap::<String, HashMap<usize, mpsc::UnboundedSender<String>>>::new()));

	let state = AppState {
		registry: registry.clone(),
		nats_client: nats_client.clone(),
	};

	// Spawn NATS wildcard subscriber task to listen for clustered broadcasts
	let nats_sub_client = nats_client.clone();
	let nats_registry = registry.clone();
	tokio::spawn(async move {
		let mut subscription = nats_sub_client
			.subscribe("strata.realtime.>")
			.await
			.expect("Failed to subscribe to NATS wildcard topic");

		info!("NATS background subscriber active on 'strata.realtime.>'");

		while let Some(message) = subscription.next().await {
			// Extract channel name from subject: e.g. "strata.realtime.chat_1" -> "chat_1"
			let subject = message.subject.as_str();
			let channel = match subject.strip_prefix("strata.realtime.") {
				Some(c) => c.to_string(),
				None => continue,
			};

			let payload = match std::str::from_utf8(&message.payload) {
				Ok(s) => s.to_string(),
				Err(_) => continue,
			};

			// Broadcast NATS payload to all local WebSockets registered to this channel
			let read_registry = nats_registry.read().await;
			if let Some(clients) = read_registry.get(&channel) {
				debug!("Broadcasting NATS message to {} local subscribers on channel {}", clients.len(), channel);
				for sender in clients.values() {
					let _ = sender.send(payload.clone());
				}
			}
		}
	});

	// Configure routing
	let app = Router::new()
		.route("/v1/realtime/health", get(health_handler))
		.route("/v1/realtime", get(ws_upgrade_handler))
		.with_state(state);

	// Address binding logic
	let bind_addr = if port.starts_with(':') {
		format!("0.0.0.0{}", port)
	} else {
		port
	};
	let listener = tokio::net::TcpListener::bind(&bind_addr)
		.await
		.unwrap_or_else(|err| panic!("Failed to bind port {}: {}", bind_addr, err));

	info!("Realtime Axum server listening on {}", bind_addr);
	axum::serve(listener, app.into_make_service_with_connect_info::<SocketAddr>())
		.await
		.expect("Failed to start Axum HTTP engine");
}

async fn health_handler() -> Json<HealthResponse> {
	Json(HealthResponse {
		status: "ok".to_string(),
		service: "realtime-rust".to_string(),
	})
}

async fn ws_upgrade_handler(
	ws: WebSocketUpgrade,
	State(state): State<AppState>,
) -> Response {
	ws.on_upgrade(move |socket| handle_websocket(socket, state))
}

async fn handle_websocket(socket: ws::WebSocket, state: AppState) {
	let (mut ws_sink, mut ws_stream) = socket.split();
	let client_id = CLIENT_ID_COUNTER.fetch_add(1, Ordering::Relaxed);

	// Client-specific channel to forward messages to WebSocket sender task
	let (tx, mut rx) = mpsc::unbounded_channel::<String>();

	// Task 1: Forward messages from internal channel to client WebSocket
	tokio::spawn(async move {
		while let Some(msg) = rx.recv().await {
			if ws_sink.send(ws::Message::Text(msg)).await.is_err() {
				break; // WebSocket connection closed
			}
		}
	});

	// Local track of rooms this socket has joined for automatic cleanup on disconnect
	let mut joined_channels = HashSet::<String>::new();

	// Task 2: Process incoming WebSocket frames from client
	while let Some(Ok(message)) = ws_stream.next().await {
		let text = match message {
			ws::Message::Text(t) => t,
			ws::Message::Binary(_) => continue, // Ignore binary
			ws::Message::Close(_) => break,     // Closed connection
			_ => continue,
		};

		// Parse dynamic JSON events
		let event: ClientEvent = match serde_json::from_str(&text) {
			Ok(e) => e,
			Err(err) => {
				let _ = tx.send(format!(r#"{{"error":"invalid_json","message":"{}"}}"#, err));
				continue;
			}
		};

		match event {
			ClientEvent::Subscribe { channel } => {
				info!("Client {} subscribing to channel '{}'", client_id, channel);
				let mut write_registry = state.registry.write().await;
				write_registry
					.entry(channel.clone())
					.or_insert_with(HashMap::new)
					.insert(client_id, tx.clone());
				joined_channels.insert(channel);
			}
			ClientEvent::Unsubscribe { channel } => {
				info!("Client {} unsubscribing from channel '{}'", client_id, channel);
				let mut write_registry = state.registry.write().await;
				if let Some(clients) = write_registry.get_mut(&channel) {
					clients.remove(&client_id);
				}
				joined_channels.remove(&channel);
			}
			ClientEvent::Broadcast { channel, data } => {
				debug!("Client {} publishing broadcast to channel '{}'", client_id, channel);
				let payload = serde_json::json!({
					"channel": channel,
					"data": data,
					"sender_id": client_id
				});
				let payload_str = payload.to_string();

				// Publish payload across cluster using NATS broker backplane
				let subject = format!("strata.realtime.{}", channel);
				state.nats_client.publish(subject, payload_str.into()).await.unwrap();
			}
		}
	}

	// Clean up client registrations on close/disconnect
	info!("WS client {} disconnected. Cleaning up memberships.", client_id);
	let mut write_registry = state.registry.write().await;
	for channel in joined_channels {
		if let Some(clients) = write_registry.get_mut(&channel) {
			clients.remove(&client_id);
		}
	}
}
