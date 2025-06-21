import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    stages: [
        { duration: '10m', target: 50 },
        { duration: '20m', target: 50 },
        { duration: '10m', target: 0 },
    ],
};

export default function () {
    http.get("http://localhost/api/question/two-sum", { headers: { Host: "localhost" } });
    sleep(1);
}
