generate-jwt-keypair:
	ssh-keygen -t rsa -b 4096 -m PEM -f test-jwt.key
	openssl rsa -in test-jwt.key -pubout -outform PEM -out test-jwt.pem
	rm test-jwt.key.pub