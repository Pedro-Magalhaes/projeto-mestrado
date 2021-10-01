from broker import Broker
from confluent_kafka import SerializingProducer

class KafkaBroker(Broker):
    # , 'queue.buffering.max.messages': 100  # test for queue impact! the number should be high
    def __init__(self, config = {'bootstrap.servers': 'localhost:9092'}) -> None:
        super().__init__()
        self._connection = SerializingProducer(config) # thread safe producer, messages are add to a buffer then are sent in batch

    ## Will be used as a callback for the async "ack" from the kafka "produce"   
    def _deliverycallback(self, kafkaerror, msg):
        if kafkaerror == None:
            pass
            #print('[INFO] - Mensagem enviada', kafkaerror, msg)
        elif(type(kafkaerror) is BufferError):
            print('[ERROR] - BUFFER!', kafkaerror) # should checkpoint and retry!
        else:
            print('[ERROR] - Erro ao enviar mensagem', kafkaerror)
    

    ## Will put in the buffer and then send async
    def send(self, topic, msg ,key = None ):
        try:            
            self._connection.produce(topic,key ,msg , on_delivery=self._deliverycallback)
        except BufferError:
            self.checkpoint(1)
            self.send(topic,msg ,key)

        # self._connection.poll(10) # FIXME: melhorar lógica de produçao de mensagens

    ## checkpoint to send the buffered messages
    def checkpoint(self, time = 1):
        self._connection.poll(time)

if 'main' in __name__:
    print('Hello')
    kb = KafkaBroker()
    for i in range(1000):
        msg = input('Fala ai: ')
        if msg == "0":
            break
        kb.send('monitor_interesse', msg)
    kb.checkpoint()

