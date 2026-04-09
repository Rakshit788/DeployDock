package internals

import (
	"fmt"
	"os"
)

func GenerateNginxConfig(subdomain string, port int) error {
	config := fmt.Sprintf(`
server {
    listen 80;
    server_name %s.vercel-clone.local;

    location / {
        proxy_pass http://host.docker.internal:%d;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}`, subdomain, port)

	filePath := fmt.Sprintf("./vhosts/%s.conf", subdomain)
	return os.WriteFile(filePath, []byte(config), 0644)
}
