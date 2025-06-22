import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
    stages: [
        { duration: '10s', target: 50 },  // ramp-up
        { duration: '20s', target: 500 }, // spike
        { duration: '10s', target: 0 },   // ramp-down
    ],
};

const questions = [
    "two-sum",
    "add-two-numbers",
    "median-of-two-sorted-arrays",
    "longest-palindromic-substring",
    "zigzag-conversion",
    "reverse-integer",
    "string-to-integer-atoi",
    "palindrome-number",
    "regular-expression-matching",
    "container-with-most-water",
];

export default function () {
    const slug = questions[Math.floor(Math.random() * questions.length)];
    const res = http.get(`http://localhost/api/question/${slug}`);

    check(res, {
        'is status 200 or pending': (r) => r.status === 200 || r.status === 202,
    });

    sleep(1);
}
