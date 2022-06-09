import { check } from 'k6';
import { writer, produce, reader, consume, createTopic } from 'k6/x/kafka';
// Arquivos:
import file from 'k6/x/file';

const bootstrapServers = ['localhost:9092'];
const kafkaTopic = 'monitor_interesse';


const producer = writer(bootstrapServers, kafkaTopic);
const consumer = reader(bootstrapServers, 'monitor_estado');
const filepath = './sample-file.txt';


export function setup() {
  console.log('Alô Mundo!');
  file.writeString(filepath, 'New file, First line.\n');
  file.appendString(filepath, 'Second line.');
}


// createTopic(bootstrapServers[0], kafkaTopic);


export default function () {
    const numberOfFiles = 3;
    let fileIndex = 0;
    
    function getNextFileIndex() {
      const index = fileIndex;
      fileIndex = (index % numberOfFiles) + 1;
      return index;
    }
  
  
  let jobId = 1;
    let currentId = jobId++;
    const messages = [
      {
        key: 'job-' + currentId,
        value: JSON.stringify({
          project: 'sample-project',
          path: 'test-file-' + getNextFileIndex() + '.txt',
          watch: true,
        }),
      },
      {
        key: 'job-' + currentId,
        value: JSON.stringify({
          project: 'sample-project',
          path: 'test-file-' + getNextFileIndex() + '.txt',
          watch: true,
        }),
      },
    ];

    messages.forEach(msg => {
      const msgValue = JSON.parse(msg.value);
      console.log(msgValue['path'])
      createTopic(bootstrapServers[0], 'test_files__' + msgValue['path']);
    });
  
    const error = produce(producer, messages);
    check(error, {
      'is sent': (err) => err == undefined,
    });

    let result = consume(consumer, 1);

    check(result, {
      "1 message returned": (msgs) => { 
        msgs.length == 1
      },
  });

    // Editar os arquivos manualmente e verificar se o teste vai receber o conteudo dos arquivos
    // podiamos fazer um programa parametrizado de escrita nos arquivos. Investigar a escrita "async", 
    // encapsular para escrever em blocos de X bytes e em determinada taxa 
    // Como vamos controlar em que "tempo" cada evento do teste vai ocorrer. (falhas, escritas, solicitações)
    // Motivação desse tipo de teste de sistema*, focar na execução e com repetir cenários de testes de sistema
    // começar a escrever a proposta: 
    // - introdução
    // - explicar o problema mostrando trabalhos relacionados (usar o survey por exemplo) 
    // - o que já foi feito

    // verificar como é implementada a API do k6 para a execução dos scripts. Executa em Go? As libs são em Go?



}
export function teardown(data) {
    producer.close();
    consumer.close();
    file.writeString(filepath, '');
}
  