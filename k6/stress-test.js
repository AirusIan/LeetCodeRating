import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '10s', target: 100 },
        { duration: '10s', target: 200 },
        { duration: '10s', target: 400 },
        { duration: '10s', target: 800 },
        { duration: '10s', target: 0 },
    ],
};

export default function () {
    http.get('http://localhost:8080/api/question/two-sum');
    sleep(1);
}
