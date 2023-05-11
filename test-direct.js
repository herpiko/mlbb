import http from 'k6/http';
import { sleep } from 'k6';

const binFile = open('/Users/herpiko/src/mlbb/golang.jpg', 'b');

export default function () {
  const data = {
    file: http.file(binFile, 'test.bin'),
  };

  const res = http.post('http://localhost:8080/upload-direct', data);
  sleep(3);
}
