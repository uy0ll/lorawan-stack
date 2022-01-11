#!/bin/bash
echo "Certificate: "
openssl x509 -noout -modulus -in client-cloud.pem |openssl sha1
echo "Key: "$2
openssl rsa -noout -modulus -in client-cloud.key |openssl sha1
