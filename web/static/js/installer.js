function testConnection() {
    const btn = document.getElementById('testBtn');
    const result = document.getElementById('testResult');

    const data = {
        host: document.getElementById('db_host').value,
        port: parseInt(document.getElementById('db_port').value) || 5432,
        user: document.getElementById('db_user').value,
        password: document.getElementById('db_password').value,
        dbname: document.getElementById('db_name').value,
        sslmode: document.getElementById('db_sslmode').value
    };

    btn.setAttribute('aria-busy', 'true');
    btn.disabled = true;
    result.className = 'test-result show alert alert-info';
    result.textContent = 'Probando conexión...';

    fetch('/install/test-connection', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        btn.removeAttribute('aria-busy');
        btn.disabled = false;

        if (data.success) {
            result.className = 'test-result show alert alert-success';
            result.textContent = 'Conexión exitosa a la base de datos.';
        } else {
            result.className = 'test-result show alert alert-error';
            result.textContent = 'Error: ' + (data.error || 'No se pudo conectar');
        }
    })
    .catch(err => {
        btn.removeAttribute('aria-busy');
        btn.disabled = false;
        result.className = 'test-result show alert alert-error';
        result.textContent = 'Error de red: ' + err.message;
    });
}
