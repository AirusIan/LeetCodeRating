import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
    vus: 1,
    duration: '10s',
};

export default function () {
    http.get("http://localhost/api/question/two-sum", { headers: { Host: "localhost" } });
    sleep(1);
}
