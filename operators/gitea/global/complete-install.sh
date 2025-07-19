#!/bin/bash
set -e

echo "Completing Gitea installation with PostgreSQL database..."

# Create a job to complete the installation
oc apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: gitea-complete-install
  namespace: gitea-system
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: gitea-install
        image: curlimages/curl:latest
        command:
        - /bin/sh
        - -c
        - |
          set -e
          echo "Waiting for Gitea to be ready..."
          until curl -f http://gitea.gitea-system.svc.cluster.local:3000/api/healthz; do
            echo "Waiting for Gitea..."
            sleep 5
          done
          
          echo "Completing Gitea installation..."
          curl -X POST http://gitea.gitea-system.svc.cluster.local:3000/install \\
            -H "Content-Type: application/x-www-form-urlencoded" \\
            -d "db_type=PostgreSQL&db_host=gitea-db.gitea-system.svc.cluster.local&db_port=5432&db_user=gitea&db_passwd=giteapassword123&db_name=gitea&ssl_mode=disable&db_schema=&charset=utf8&app_name=Gitea%3A+Git+with+a+cup+of+tea&repo_root_path=%2Ftmp%2Fgitea%2Frepositories&lfs_root_path=%2Ftmp%2Fgitea%2Fdata%2Flfs&run_user=gitea&domain=gitea.gitea-system.svc.cluster.local&ssh_port=22&http_port=3000&app_url=http%3A%2F%2Fgitea.gitea-system.svc.cluster.local%3A3000%2F&log_root_path=%2Ftmp%2Fgitea%2Flog&smtp_addr=&smtp_port=587&smtp_from=&smtp_user=&smtp_passwd=&enable_federated_avatar=on&enable_open_id_sign_in=on&enable_open_id_sign_up=on&default_allow_create_organization=on&default_enable_timetracking=on&no_reply_address=noreply.localhost&password_algorithm=pbkdf2&admin_name=admin&admin_passwd=bootstrap123&admin_confirm_passwd=bootstrap123&admin_email=admin%40bootstrap.local"
          
          sleep 5
          
          echo "Testing API access..."
          curl -f http://gitea.gitea-system.svc.cluster.local:3000/api/v1/version || echo "API not ready yet"
          
          echo "Creating bootstrap organization..."
          curl -X POST http://gitea.gitea-system.svc.cluster.local:3000/api/v1/orgs \\
            -H "Content-Type: application/json" \\
            -u "admin:bootstrap123" \\
            -d '{
              "username": "bootstrap",
              "full_name": "Bootstrap Organization",
              "description": "Organization for bootstrap repositories"
            }' || echo "Organization creation failed or already exists"
          
          echo "Creating bootstrap repository..."
          curl -X POST http://gitea.gitea-system.svc.cluster.local:3000/api/v1/orgs/bootstrap/repos \\
            -H "Content-Type: application/json" \\
            -u "admin:bootstrap123" \\
            -d '{
              "name": "bootstrap", 
              "description": "Bootstrap cluster configurations",
              "private": false,
              "auto_init": true
            }' || echo "Repository creation failed or already exists"
          
          echo "✅ Gitea installation and setup complete!"
EOF

echo "Waiting for installation to complete..."
oc wait --for=condition=complete job/gitea-complete-install -n gitea-system --timeout=300s

echo "✅ Gitea installation completed successfully!"