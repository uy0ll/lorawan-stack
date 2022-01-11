openssl s_client -cert client-cloud.pem -key client-cloud.key -CAfile ca-cloud.pem \
 -showcerts -connect server.lorawan.co.ua:8886
