import { check } from 'k6';
import exec from 'k6/execution';
import { writer, produce, createTopic } from 'k6/x/kafka';
import file from 'k6/x/file';

const bootstrapServers = ['localhost:9092'];
const kafkaTopic = 'teste_ordem';


const producer = writer(bootstrapServers, kafkaTopic);


export function setup() {
    createTopic(bootstrapServers[0], kafkaTopic);
}

function mockRequest(vuId) {
    return  new Promise((resolve, reject) => {
        // setTimeout(() => {
        resolve(JSON.stringify({id: vuId}));
        // }, 1000);
    });
}

// async function getMsg(vuId) {
//     return await mockRequest(vuId);
// }

function getMsg(vuId) {
    return mockRequest(vuId);
}


export default function () {
    getMsg(exec.vu.idInTest).then(
        (message) => {
            const error = produce(producer, message);
            check(error, {
            'is sent': (err) => err == undefined,
            });
        }
    )    
}


  