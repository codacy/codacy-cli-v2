

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