# HttpDump

## Usage Examples
* Basic usage hosts an open HTTP port on 0.0.0.0:9999 that responds with HTTP 200 to anything is receives
** HTTPDump -response
* Host on port 80 with custom JSON response
** HTTPDump -port 80 -response '{"test": false}'
* Return redirect to https://google.com
** ** HTTPDump -port 80 -redirect 'https://google.com'
* Run TLS on port 443. Return HTTP 500
** HTTPDump -port 443 -tls -tls-key ./key.pem -tls-cert ./certificate.pem -response-code 500 -response "NOOOOOOOOOOOPE"
