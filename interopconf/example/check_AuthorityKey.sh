echo "Verify ca.pem <-> cert.pem"
openssl verify -CAfile ca-cloud.pem client-cloud.pem
echo ""
echo "Check AuthorityKey"
openssl x509 -in client-cloud.pem -text -noout | grep -A 1 Identifier; openssl x509 -in ca-cloud.pem -text -noout | grep -A 1 Identifier
