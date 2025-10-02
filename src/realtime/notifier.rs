use tokio::sync::broadcast;

#[derive(Clone)]
pub struct Notifier {
    pub sender: broadcast::Sender<String>,
}

pub fn create_notifier() -> Notifier {
    let (tx, _rx) = broadcast::channel(1024);
    Notifier { sender: tx }
}

impl Notifier {
    pub fn subscribe(&self) -> broadcast::Receiver<String> { self.sender.subscribe() }
    pub fn broadcast(&self, text: &str) { let _ = self.sender.send(text.to_string()); }
}

