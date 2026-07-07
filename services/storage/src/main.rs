use axum::{
	extract::{Multipart, Path, Query, State},
	http::{header, StatusCode},
	response::{IntoResponse, Response},
	routing::{delete, get, post},
	Json, Router,
};
use image::{ImageFormat, ImageOutputFormat};
use serde::{Deserialize, Serialize};
use std::{net::SocketAddr, sync::Arc};
use s3::{bucket::Bucket, creds::Credentials, BucketConfiguration, Region};
use tokio::sync::Mutex;
use tracing::{debug, error, info};

#[derive(Clone)]
struct AppConfig {
	endpoint: String,
	access_key: String,
	secret_key: String,
}

#[derive(Clone)]
struct AppState {
	config: AppConfig,
	buckets: Arc<Mutex<Vec<BucketResponse>>>,
}

#[derive(Deserialize)]
struct ResizeParams {
	width: Option<u32>,
	height: Option<u32>,
}

#[derive(Serialize, Clone, Debug)]
struct BucketResponse {
	name: String,
	created_at: String,
}

#[derive(Deserialize)]
struct CreateBucketRequest {
	name: String,
}

#[derive(Serialize)]
struct UploadResponse {
	filepath: String,
	url: String,
	size: usize,
}

#[tokio::main]
async fn main() {
	tracing_subscriber::fmt()
		.with_env_filter(
			tracing_subscriber::EnvFilter::try_from_default_env()
				.unwrap_or_else(|_| "storage=debug,axum=info".into()),
		)
		.init();

	info!("Starting NovaBase Storage Service in Rust...");

	let port = std::env::var("PORT").unwrap_or_else(|_| ":8084".to_string());
	let config = AppConfig {
		endpoint: std::env::var("MINIO_ENDPOINT").unwrap_or_else(|_| "http://minio:9000".to_string()),
		access_key: std::env::var("MINIO_ACCESS_KEY").unwrap_or_else(|_| "novabase_storage_admin".to_string()),
		secret_key: std::env::var("MINIO_SECRET_KEY").unwrap_or_else(|_| "minio_secure_pass_123".to_string()),
	};

	// Initialize the in-memory bucket registry with default provisioned buckets
	let now = chrono::Utc::now().to_rfc3339();
	let initial_buckets = vec![
		BucketResponse {
			name: "novabase-storage".to_string(),
			created_at: now.clone(),
		},
		BucketResponse {
			name: "novabase-functions".to_string(),
			created_at: now,
		},
	];

	let state = AppState {
		config,
		buckets: Arc::new(Mutex::new(initial_buckets)),
	};

	let app = Router::new()
		.route("/v1/storage/health", get(health_handler))
		.route("/v1/storage/buckets", get(list_buckets).post(create_bucket))
		.route("/v1/storage/buckets/:bucket", delete(delete_bucket))
		.route("/v1/storage/buckets/:bucket/upload", post(upload_file))
		.route("/v1/storage/buckets/:bucket/download/*filepath", get(download_file))
		.with_state(state);

	let bind_addr = if port.starts_with(':') {
		format!("0.0.0.0{}", port)
	} else {
		port
	};

	let listener = tokio::net::TcpListener::bind(&bind_addr)
		.await
		.unwrap_or_else(|err| panic!("Failed to bind port {}: {}", bind_addr, err));

	info!("Storage Axum server listening on {}", bind_addr);
	axum::serve(listener, app.into_make_service_with_connect_info::<SocketAddr>())
		.await
		.expect("Failed to start Axum HTTP engine");
}

async fn health_handler() -> impl IntoResponse {
	(StatusCode::OK, Json(serde_json::json!({
		"status": "ok",
		"service": "storage-rust"
	})))
}

fn get_s3_bucket(bucket_name: &str, config: &AppConfig) -> Result<Bucket, Response> {
	let credentials = Credentials::new(
		Some(&config.access_key),
		Some(&config.secret_key),
		None,
		None,
		None,
	)
	.map_err(|err| {
		error!("Failed to build S3 credentials: {}", err);
		(StatusCode::INTERNAL_SERVER_ERROR, "S3 Configuration Error").into_response()
	})?;

	let region = Region::Custom {
		region: "us-east-1".to_string(),
		endpoint: config.endpoint.clone(),
	};

	let bucket = Bucket::new(bucket_name, region, credentials)
		.map_err(|err| {
			error!("Failed to create bucket instance: {}", err);
			(StatusCode::INTERNAL_SERVER_ERROR, "S3 Client Error").into_response()
		})?
		.with_path_style();

	Ok(bucket)
}

async fn list_buckets(State(state): State<AppState>) -> impl IntoResponse {
	let list = state.buckets.lock().await;
	(StatusCode::OK, Json(list.clone()))
}

async fn create_bucket(
	State(state): State<AppState>,
	Json(payload): Json<CreateBucketRequest>,
) -> Result<impl IntoResponse, Response> {
	info!("Creating S3 bucket: {}", payload.name);

	let credentials = Credentials::new(
		Some(&state.config.access_key),
		Some(&state.config.secret_key),
		None,
		None,
		None,
	)
	.map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response())?;

	let region = Region::Custom {
		region: "us-east-1".to_string(),
		endpoint: state.config.endpoint.clone(),
	};

	// Create bucket on MinIO
	Bucket::create_with_path_style(
		&payload.name,
		region,
		credentials,
		BucketConfiguration::default(),
	)
	.await
	.map_err(|e| {
		error!("Failed to create S3 bucket: {}", e);
		(StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response()
	})?;

	// Update local in-memory registry
	let mut list = state.buckets.lock().await;
	list.push(BucketResponse {
		name: payload.name.clone(),
		created_at: chrono::Utc::now().to_rfc3339(),
	});

	Ok((StatusCode::CREATED, Json(serde_json::json!({
		"message": "Bucket successfully created",
		"bucket": payload.name
	}))))
}

async fn delete_bucket(
	Path(bucket_name): Path<String>,
	State(state): State<AppState>,
) -> Result<impl IntoResponse, Response> {
	info!("Deleting S3 bucket: {}", bucket_name);
	let bucket = get_s3_bucket(&bucket_name, &state.config)?;

	// Delete bucket on MinIO
	bucket.delete().await.map_err(|e| {
		error!("Failed to delete S3 bucket: {}", e);
		(StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response()
	})?;

	// Update local in-memory registry
	let mut list = state.buckets.lock().await;
	list.retain(|b| b.name != bucket_name);

	Ok((StatusCode::OK, Json(serde_json::json!({
		"message": "Bucket successfully deleted"
	}))))
}

async fn upload_file(
	Path(bucket_name): Path<String>,
	State(state): State<AppState>,
	mut multipart: Multipart,
) -> Result<impl IntoResponse, Response> {
	let bucket = get_s3_bucket(&bucket_name, &state.config)?;

	let mut filepath = String::new();
	let mut data = Vec::new();
	let mut content_type = "application/octet-stream".to_string();

	while let Some(field) = multipart
		.next_field()
		.await
		.map_err(|e| (StatusCode::BAD_REQUEST, e.to_string()).into_response())?
	{
		let name = field.name().unwrap_or("").to_string();
		if name == "file" {
			filepath = field.file_name().unwrap_or("unnamed_file").to_string();
			if let Some(ct) = field.content_type() {
				content_type = ct.to_string();
			}
			data = field
				.bytes()
				.await
				.map_err(|e| (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response())?
				.to_vec();
			break;
		}
	}

	if filepath.is_empty() || data.is_empty() {
		return Err((StatusCode::BAD_REQUEST, "File upload missing 'file' field or payload is empty").into_response());
	}

	info!("Uploading file '{}' to bucket '{}' ({} bytes)", filepath, bucket_name, data.len());
	bucket
		.put_object_with_content_type(&filepath, &data, &content_type)
		.await
		.map_err(|e| {
			error!("MinIO upload failed: {}", e);
			(StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response()
		})?;

	let file_url = format!("/v1/storage/buckets/{}/download/{}", bucket_name, filepath);

	Ok((
		StatusCode::CREATED,
		Json(UploadResponse {
			filepath,
			url: file_url,
			size: data.len(),
		}),
	))
}

async fn download_file(
	Path((bucket_name, filepath)): Path<(String, String)>,
	Query(params): Query<ResizeParams>,
	State(state): State<AppState>,
) -> Result<impl IntoResponse, Response> {
	let bucket = get_s3_bucket(&bucket_name, &state.config)?;

	// Read object from MinIO S3
	let response_data = bucket.get_object(&filepath).await.map_err(|e| {
		error!("S3 object download failed: {}", e);
		(StatusCode::NOT_FOUND, "File not found").into_response()
	})?;

	let mut bytes = response_data.to_vec();
	let mut resp_content_type = guess_content_type(&filepath);

	// On-the-fly image manipulation check
	if (params.width.is_some() || params.height.is_some()) && is_image_content_type(&resp_content_type) {
		debug!("Resizing image file '{}' dynamically", filepath);
		
		let img_format = match ImageFormat::from_path(&filepath) {
			Ok(fmt) => fmt,
			Err(_) => ImageFormat::Png, // Default fallback
		};

		// Load image in memory
		if let Ok(img) = image::load_from_memory_with_format(&bytes, img_format) {
			let w = params.width.unwrap_or(img.width());
			let h = params.height.unwrap_or(img.height());

			// Perform resize operation using Lanczos3 filter
			let resized = img.resize(w, h, image::imageops::FilterType::Lanczos3);

			let mut buffer = std::io::Cursor::new(Vec::new());
			let output_format = match img_format {
				ImageFormat::Jpeg => ImageOutputFormat::Jpeg(85),
				_ => ImageOutputFormat::Png,
			};

			if resized.write_to(&mut buffer, output_format).is_ok() {
				bytes = buffer.into_inner();
				if img_format == ImageFormat::Jpeg {
					resp_content_type = "image/jpeg".to_string();
				} else {
					resp_content_type = "image/png".to_string();
				}
			}
		}
	}

	Ok(Response::builder()
		.status(StatusCode::OK)
		.header(header::CONTENT_TYPE, resp_content_type)
		.header(header::CACHE_CONTROL, "public, max-age=31536000")
		.body(axum::body::Body::from(bytes))
		.unwrap())
}

fn guess_content_type(path: &str) -> String {
	let lowercase = path.to_lowercase();
	if lowercase.ends_with(".png") {
		"image/png".to_string()
	} else if lowercase.ends_with(".jpg") || lowercase.ends_with(".jpeg") {
		"image/jpeg".to_string()
	} else if lowercase.ends_with(".gif") {
		"image/gif".to_string()
	} else if lowercase.ends_with(".svg") {
		"image/svg+xml".to_string()
	} else if lowercase.ends_with(".json") {
		"application/json".to_string()
	} else if lowercase.ends_with(".pdf") {
		"application/pdf".to_string()
	} else {
		"application/octet-stream".to_string()
	}
}

fn is_image_content_type(ct: &str) -> bool {
	ct.starts_with("image/")
}
