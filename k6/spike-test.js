import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    stages: [
        { duration: '5s', target: 500 },
        { duration: '10s', target: 500 },
        { duration: '5s', target: 0 },
    ],
};

export default function () {
    http.get("http://localhost/api/question/two-sum", { headers: { Host: "localhost" } });
    sleep(1);
}
