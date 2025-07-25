#!/bin/bash
set -e

GITEA_URL="http://gitea.gitea-system.svc.cluster.local:3000"
ADMIN_USER="mturansk"
ADMIN_PASS="acmeprototype321#"
ADMIN_EMAIL="mturansk@redhat.com"

echo "Initializing Gitea repository and admin user..."

# Create a job to initialize Gitea from inside the cluster
oc apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: gitea-init3
  namespace: gitea-system
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: gitea-init
        image: curlimages/curl:latest
        command:
        - /bin/sh
        - -c
        - |
          set -e
          echo "Waiting for Gitea to be ready..."
          until curl -f \${GITEA_URL}/api/healthz; do
            echo "Waiting for Gitea..."
            sleep 5
          done
          
          echo "Gitea is ready, starting initialization..."
          
          # Try to access API first
          if curl -f \${GITEA_URL}/api/v1/version; then
            echo "Gitea API is ready"
          else
            echo "Gitea needs installation, performing setup..."
            
            # Complete installation via form submission
            curl -X POST \${GITEA_URL}/install \\
              -H "Content-Type: application/x-www-form-urlencoded" \\
              -d "db_type=SQLite3&db_path=%2Ftmp%2Fgitea%2Fgitea.db&app_name=Gitea&repo_root_path=%2Ftmp%2Fgitea%2Frepositories&domain=gitea.gitea-system.svc.cluster.local&ssh_port=22&http_port=3000&app_url=http%3A%2F%2Fgitea.gitea-system.svc.cluster.local%3A3000%2F&admin_name=\${ADMIN_USER}&admin_passwd=\${ADMIN_PASS}&admin_confirm_passwd=\${ADMIN_PASS}&admin_email=\${ADMIN_EMAIL}" || true
            
            # Wait a bit for installation to complete
            sleep 10
          fi
          
          # Create bootstrap organization
          echo "Creating bootstrap organization..."
          curl -X POST \${GITEA_URL}/api/v1/orgs \\
            -H "Content-Type: application/json" \\
            -u "\${ADMIN_USER}:\${ADMIN_PASS}" \\
            -d '{
              "username": "bootstrap",
              "full_name": "Bootstrap Organization",
              "description": "Organization for bootstrap repositories"
            }' || true
          
          # Create bootstrap repository
          echo "Creating bootstrap repository..."
          curl -X POST \${GITEA_URL}/api/v1/orgs/bootstrap/repos \\
            -H "Content-Type: application/json" \\
            -u "\${ADMIN_USER}:\${ADMIN_PASS}" \\
            -d '{
              "name": "bootstrap",
              "description": "Bootstrap cluster configurations",
              "private": false,
              "auto_init": true
            }' || true
          
          echo "Gitea initialization complete!"
        env:
        - name: GITEA_URL
          value: "${GITEA_URL}"
        - name: ADMIN_USER
          value: "${ADMIN_USER}"
        - name: ADMIN_PASS
          value: "${ADMIN_PASS}"
        - name: ADMIN_EMAIL
          value: "${ADMIN_EMAIL}"
EOF

echo "Gitea initialization job created. Monitoring progress..."
oc wait --for=condition=complete job/gitea-init -n gitea-system --timeout=300s

echo "Gitea setup completed successfully!"
echo "Repository URL: ${GITEA_URL}/bootstrap/bootstrap"
echo "Admin credentials: ${ADMIN_USER}/${ADMIN_PASS}"