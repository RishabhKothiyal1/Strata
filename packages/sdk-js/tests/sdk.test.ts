import { Client } from 'pg';
import { execSync } from 'child_process';
import { NovaBaseClient } from '../src';
import * as assert from 'assert';

const GATEWAY_URL = 'http://localhost:8000';
const PG_URL = 'postgres://novabase_admin:novabase_secure_pass_123@localhost:5432/novabase';

async function runTests() {
  console.log('🚀 Initializing NovaBase Integration Test Suite...\n');

  // 1. Setup Test Database Table
  const pgClient = new Client({ connectionString: PG_URL });
  await pgClient.connect();
  console.log('📦 Connected to PostgreSQL. Creating test table public.sdk_todos...');
  await pgClient.query(`
    CREATE TABLE IF NOT EXISTS public.sdk_todos (
      id SERIAL PRIMARY KEY,
      title VARCHAR(255) NOT NULL,
      completed BOOLEAN NOT NULL DEFAULT false,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
  `);

  console.log('🔄 Restarting REST service to trigger schema introspection...');
  execSync('docker compose restart rest', { stdio: 'inherit' });
  
  // Wait a bit for rest service to compile / restart and healthcheck to pass
  console.log('⏳ Waiting 5 seconds for REST service to boot...');
  await new Promise((r) => setTimeout(r, 5000));

  // Initialize client
  const client = new NovaBaseClient(GATEWAY_URL);

  const email = `sdk-test-${Date.now()}@example.com`;
  const password = 'secure_password_123';
  let userId: any = null;

  try {
    // 2. Authentication Tests
    console.log('\n--- 🔑 1. Authentication Tests ---');
    console.log(`Signing up user: ${email}...`);
    const signupUser = await client.auth.signUp(email, password);
    assert.ok(signupUser.id, 'Signup should return user ID');
    assert.strictEqual(signupUser.email, email, 'Signup email mismatch');
    console.log('✅ Signup Successful.');

    console.log('Signing in...');
    const loginResult = await client.auth.signIn(email, password);
    assert.ok(loginResult.session.access_token, 'Login should yield JWT access token');
    userId = loginResult.user.id;
    console.log(`✅ Login Successful. Active User ID: ${userId}`);

    // 3. Dynamic REST CRUD Tests
    console.log('\n--- 🗄️ 2. Dynamic REST CRUD Tests ---');
    console.log('Inserting todo record...');
    const insertResult: any = await client.from('sdk_todos').insert({
      title: 'Write SDK tests',
      completed: false
    }).execute();
    assert.ok(insertResult.id, 'Insert should return inserted row ID');
    console.log(`✅ Todo inserted. ID: ${insertResult.id}`);

    console.log('Selecting todo record with filters...');
    const selectResult: any = await client.from('sdk_todos')
      .select('*')
      .eq('title', 'Write SDK tests')
      .execute();
    assert.strictEqual(selectResult.length, 1, 'Select should return 1 record');
    assert.strictEqual(selectResult[0].completed, false, 'Todo should be incomplete');
    console.log('✅ Select filters verified.');

    console.log('Updating todo record...');
    const updateResult: any = await client.from('sdk_todos')
      .update({ completed: true })
      .eq('id', insertResult.id)
      .execute();
    assert.strictEqual(updateResult.rows_affected, 1, 'Update should affect 1 row');

    const selectResult2: any = await client.from('sdk_todos')
      .select('*')
      .eq('id', insertResult.id)
      .execute();
    assert.strictEqual(selectResult2[0].completed, true, 'Updated todo should be completed');
    console.log('✅ Update verified.');

    console.log('Deleting todo record...');
    const deleteResult: any = await client.from('sdk_todos')
      .delete()
      .eq('id', insertResult.id)
      .execute();
    assert.strictEqual(deleteResult.rows_affected, 1, 'Delete should affect 1 row');
    console.log('✅ Delete verified.');

    // 4. Storage Bucket & File Tests
    console.log('\n--- 🪣 3. Storage Bucket & File Tests ---');
    const bucketName = 'sdk-test-bucket';
    console.log(`Creating storage bucket: ${bucketName}...`);
    const createBucketRes = await client.storage.createBucket(bucketName);
    console.log(`✅ Bucket create message: ${createBucketRes.message}`);

    console.log('Uploading sample text file to bucket...');
    const uploadRes = await client.storage.from(bucketName).upload(
      'notes.txt',
      Buffer.from('Welcome to NovaBase SDK storage!'),
      'text/plain'
    );
    assert.strictEqual(uploadRes.filepath, 'notes.txt', 'Uploaded file path mismatch');
    console.log(`✅ Upload completed. File URL: ${uploadRes.url}`);

    console.log('Downloading uploaded file...');
    const dlResponse = await client.storage.from(bucketName).download('notes.txt');
    const dlText = await dlResponse.text();
    assert.strictEqual(dlText, 'Welcome to NovaBase SDK storage!', 'Downloaded content mismatch');
    console.log('✅ Download content verified.');

    console.log('Deleting test bucket...');
    const deleteBucketRes = await client.storage.deleteBucket(bucketName);
    console.log(`✅ Bucket delete message: ${deleteBucketRes.message}`);

    // 5. Functions Engine Tests
    console.log('\n--- ⚡ 4. Functions Engine Tests ---');
    const fnName = 'sdk-echo';
    const fnCode = `
      function handler(request) {
        return { statusCode: 200, body: { greeting: "Hello " + request.body.name } };
      }
    `;
    console.log(`Deploying JavaScript function: ${fnName}...`);
    const deployRes = await client.functions.deploy(fnName, fnCode, 'SDK test echo function');
    assert.strictEqual(deployRes.name, fnName, 'Deployed function name mismatch');
    console.log('✅ Function deployed successfully.');

    console.log('Invoking deployed function...');
    const invokeRes = await client.functions.invoke(fnName, { name: 'Superstar' });
    assert.strictEqual(invokeRes.status_code, 200, 'Function status code should be 200');
    assert.strictEqual(invokeRes.body.greeting, 'Hello Superstar', 'Returned body greeting mismatch');
    console.log(`✅ Invocation Successful. Response: ${JSON.stringify(invokeRes.body)}`);

    console.log('Deleting function...');
    const deleteFnRes = await client.functions.delete(fnName);
    console.log(`✅ Function delete message: ${deleteFnRes.message}`);

    // 6. AI & Semantic Search Tests
    console.log('\n--- 🧠 5. AI & Semantic Search Tests ---');
    const collName = 'sdk-search-kb';
    console.log(`Creating AI Vector collection: ${collName}...`);
    const createCollRes = await client.ai.createCollection(collName, 'SDK test collection');
    assert.strictEqual(createCollRes.name, collName, 'Collection name mismatch');

    console.log('Indexing semantic documents...');
    await client.ai.collection(collName).addDocument(
      'The storage service resizes images using Lanczos3 filter dynamically.',
      { topic: 'storage' }
    );
    await client.ai.collection(collName).addDocument(
      'Authentication service handles bcrypt hashing and secure JWT session issuance.',
      { topic: 'auth' }
    );
    console.log('✅ Ingestion complete.');

    console.log('Performing semantic search...');
    const searchRes = await client.ai.collection(collName).search('how do we resize images', 1);
    assert.strictEqual(searchRes.results.length, 1, 'Search should return 1 hit');
    assert.strictEqual(searchRes.results[0].metadata.topic, 'storage', 'Top hit topic should be storage');
    console.log(`✅ Search result top hit: "${searchRes.results[0].content}" (Score: ${searchRes.results[0].score.toFixed(4)})`);

    console.log('Deleting AI collection...');
    const deleteCollRes = await client.ai.deleteCollection(collName);
    console.log(`✅ Collection delete message: ${deleteCollRes.message}`);

    // 7. Realtime WebSockets Tests
    console.log('\n--- 📡 6. Realtime WebSockets Tests ---');
    const channelName = 'sdk-chat-room';
    const realtimeMsg = { text: 'Hello, is this working?' };
    let receivedPayload: any = null;

    console.log(`Subscribing to WebSocket channel: ${channelName}...`);
    const subscription = client.realtime.channel(channelName).subscribe((payload) => {
      console.log('📥 Received real-time broadcast payload!');
      receivedPayload = payload;
    });

    console.log('Waiting 1 second for WebSocket handshake...');
    await new Promise((r) => setTimeout(r, 1000));

    console.log('Broadcasting test message...');
    client.realtime.channel(channelName).broadcast(realtimeMsg);

    console.log('Waiting 2 seconds for WebSocket reception...');
    await new Promise((r) => setTimeout(r, 2000));

    assert.ok(receivedPayload, 'Should have received broadcast message');
    assert.deepStrictEqual(receivedPayload.data, realtimeMsg, 'Received message data mismatch');
    console.log('✅ Real-time WS message broadcast & loopback verified.');

    subscription.unsubscribe();
    client.realtime.disconnect();
    console.log('✅ Realtime unsubscribed and disconnected.');

    console.log('\n🎉 ALL SDK INTEGRATION TESTS PASSED SUCCESSFULLY! 🎉\n');

  } catch (error) {
    console.error('❌ Test execution failed with error:', error);
    process.exitCode = 1;
  } finally {
    // 8. Cleanup Database Table & Session
    console.log('🧹 Cleaning up test table public.sdk_todos...');
    await pgClient.query('DROP TABLE IF EXISTS public.sdk_todos;');
    await pgClient.end();

    if (userId) {
      console.log('🧹 Cleaning up test user from DB...');
      const cleanPgClient = new Client({ connectionString: PG_URL });
      await cleanPgClient.connect();
      await cleanPgClient.query('DELETE FROM public.users WHERE email = $1', [email]);
      await cleanPgClient.end();
    }

    console.log('🔄 Restarting REST service to clean up introspected tables...');
    execSync('docker compose restart rest', { stdio: 'inherit' });
  }
}

runTests();
