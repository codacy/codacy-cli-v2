// vulneravel.ts

import * as http from 'http';
import * as url from 'url';
import * as fs from 'fs';
import * as mysql from 'mysql';

// 1. Exposição de credenciais
const db = mysql.createConnection({
  host: 'localhost',
  user: 'root',
  password: 'rootpassword', // Credenciais expostas
  database: 'test'
});

http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url || '', true);
  const query = parsedUrl.query;

  // 2. Injeção SQL - FIXED: Use parameterized query
  const username = query.username;
  const sql = `SELECT * FROM users WHERE username = ?`;
  db.query(sql, [username], (err, result) => {
    if (err) throw err;

    // 3. Exposição de dados sensíveis - FIXED: Filter sensitive fields
    const safeResult = result.map((user: any) => {
      const { password, ...safeUser } = user;
      return safeUser;
    });
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify(safeResult)); // devolve dados sem campos sensíveis
  });

  // 4. Leitura insegura de ficheiros
  const file = query.file as string;
  fs.readFile('./uploads/' + file, 'utf8', (err, data) => {
    if (!err) {
      res.write('\n\n' + data); // pode ser usado para leitura arbitrária de ficheiros
    }
  });

  // 5. Execução de código arbitrário
  if (query.runCode) {
    eval(query.runCode as string); // MUITO perigoso
  }

}).listen(8080);

// 6. Dependência desatualizada (suponha que mysql está vulnerável)

// 7. Falta de HTTPS (http em vez de https)

// 8. Nenhuma validação de entrada em nenhuma parte

// 9. Stack traces revelados com throw err

// 10. Não existe autenticação nem controlo de acessos

console.log('Servidor inseguro a correr em http://localhost:8080');
