echo "Verify ca.pem <-> cert.pem"
openssl verify -CAfile ca.pem cert.pem
echo ""
echo "Check AuthorityKey"
openssl x509 -in cert.pem -text -noout | grep -A 1 Identifier; openssl x509 -in ca.pem -text -noout | grep -A 1 Identifier
