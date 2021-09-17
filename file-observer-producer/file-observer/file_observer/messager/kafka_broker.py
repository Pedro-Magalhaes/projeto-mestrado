from broker import Broker
from confluent_kafka import SerializingProducer

class KafkaBroker(Broker):
    
    def __init__(self) -> None:
        super().__init__()
        self.connection = SerializingProducer()
        self.connection.produce()

    

    ## Will put in the buffer and then send async
    def send(self, msg) -> bool:
        self.connection.send(msg)