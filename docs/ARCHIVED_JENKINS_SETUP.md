# Jenkins Setup Guide for Ubuntu 24.04 VPS

Since you are deploying Jenkins on the same VPS as your application "for now", we will run it as a service and proxy it behind Nginx.

## 1. Install Java (Required)
Jenkins requires Java to run.

```bash
sudo apt update
sudo apt install -y fontconfig openjdk-17-jre
java -version
```

## 2. Install Jenkins
```bash
sudo wget -O /usr/share/keyrings/jenkins-keyring.asc \
  https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key
echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]" \
  https://pkg.jenkins.io/debian-stable binary/ | sudo tee \
  /etc/apt/sources.list.d/jenkins.list > /dev/null
sudo apt-get update
sudo apt-get install -y jenkins
```

## 3. Enable & Start Jenkins
```bash
sudo systemctl enable jenkins
sudo systemctl start jenkins
```

## 4. Install GitHub Integration Plugin (Do this in UI)
1.  **Get Initial Password:**
    ```bash
    sudo cat /var/lib/jenkins/secrets/initialAdminPassword
    ```
2.  Open your browser to `http://<YOUR_VPS_IP>:8080`.
3.  Paste the password.
4.  Select **"Install suggested plugins"**.
5.  Create your Admin User.
6.  Go to **Manage Jenkins** -> **Plugins** -> **Available Plugins**.
7.  Search for and install:
    *   **Docker Pipeline**
    *   **GitHub Integration**
    *   **Generic Webhook Trigger**

## 5. Configure Nginx Reverse Proxy (Optional but Recommended)
To access Jenkins via `jenkins.yourdomain.com` instead of `:8080`.

1.  Create Nginx config: `sudo nano /etc/nginx/sites-available/jenkins`
    ```nginx
    server {
        listen 80;
        server_name jenkins.madabank.art; # Replace with your domain

        location / {
            proxy_pass http://localhost:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
    ```
2.  Enable site:
    ```bash
    sudo ln -s /etc/nginx/sites-available/jenkins /etc/nginx/sites-enabled/
    sudo nginx -t
    sudo systemctl restart nginx
    ```

## 6. Docker Permission for Jenkins
Jenkins needs to run Docker commands to build your app.
```bash
sudo usermod -aG docker jenkins
sudo systemctl restart jenkins
```
