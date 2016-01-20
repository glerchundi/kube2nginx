```
$> cat <<EOF | base64
{
    "items": [{
        "spec": {
            "rules": [{
                "host": "test.mydomain.org",
                "http": {
                    "paths": [{
                        "path": "/",
                        "backend": {
                            "serviceName": "test-service",
                            "servicePort": 8080
                        }
                    }]
                }
            }]
        }
    }]
}
EOF
```