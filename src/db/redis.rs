use redis::Client;

pub fn create_redis_client() -> anyhow::Result<Client> {
    let url = std::env::var("REDIS_URL").expect("REDIS_URL is required");
    let client = Client::open(url)?;
    Ok(client)
}

pub fn stream_namespace() -> String {
    std::env::var("REDIS_STREAM_NAMESPACE").unwrap_or_else(|_| "valpago:transactions".into())
}

pub fn consumer_group() -> String {
    std::env::var("REDIS_CONSUMER_GROUP").unwrap_or_else(|_| "valpago:cg".into())
}

pub fn consumer_name() -> String {
    std::env::var("REDIS_CONSUMER_NAME").unwrap_or_else(|_| "worker-1".into())
}

