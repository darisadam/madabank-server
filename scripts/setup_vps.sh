#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}===============================================${NC}"
echo -e "${GREEN}    MadaBank VPS Initialization Script         ${NC}"
echo -e "${GREEN}    Target OS: Ubuntu 24.04 LTS                ${NC}"
echo -e "${GREEN}===============================================${NC}"

# Check for root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 1: System Update & Basic Tools${NC}"
apt update && apt upgrade -y
apt install -y curl wget git unzip htop vi nano net-tools ufw fail2ban

echo -e "\n${YELLOW}Step 2: Create Admin User${NC}"
read -p "Enter new admin username (default: admin): " ADMIN_USER
ADMIN_USER=${ADMIN_USER:-admin}

if id "$ADMIN_USER" &>/dev/null; then
    echo "User $ADMIN_USER already exists."
else
    adduser "$ADMIN_USER"
    usermod -aG sudo "$ADMIN_USER"
    echo -e "${GREEN}User $ADMIN_USER created and added to sudo group.${NC}"
fi

echo -e "\n${YELLOW}Step 3: SSH Hardening${NC}"
# Copy root SSH keys to new user if they exist
if [ -d "/root/.ssh" ]; then
    echo "Copying root SSH keys to $ADMIN_USER..."
    mkdir -p /home/$ADMIN_USER/.ssh
    cp /root/.ssh/authorized_keys /home/$ADMIN_USER/.ssh/
    chown -R $ADMIN_USER:$ADMIN_USER /home/$ADMIN_USER/.ssh
    chmod 700 /home/$ADMIN_USER/.ssh
    chmod 600 /home/$ADMIN_USER/.ssh/authorized_keys
else
    echo -e "${RED}Warning: No /root/.ssh found. Make sure to setup SSH keys for $ADMIN_USER manually!${NC}"
fi

# Configure SSHD
BLOCK_ROOT_LOGIN=false
read -p "Disable Root Login? (y/n): " DISABLE_ROOT
if [[ "$DISABLE_ROOT" =~ ^[Yy]$ ]]; then
    sed -i 's/^PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
    BLOCK_ROOT_LOGIN=true
fi

read -p "Disable Password Authentication? (y/n): " DISABLE_PASS
if [[ "$DISABLE_PASS" =~ ^[Yy]$ ]]; then
    sed -i 's/^PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
fi

systemctl restart ssh
echo -e "${GREEN}SSH configuration updated.${NC}"

echo -e "\n${YELLOW}Step 4: Firewall Setup (UFW)${NC}"
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp

read -p "Enable Firewall now? (y/n): " ENABLE_FW
if [[ "$ENABLE_FW" =~ ^[Yy]$ ]]; then
    ufw --force enable
    echo -e "${GREEN}Firewall enabled.${NC}"
fi

echo -e "\n${YELLOW}Step 5: Docker Installation${NC}"
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    
    # Configure logging
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
EOF
    systemctl restart docker
else
    echo "Docker already installed."
fi

# Add user to docker group
usermod -aG docker "$ADMIN_USER"
echo -e "${GREEN}Added $ADMIN_USER to docker group.${NC}"

echo -e "\n${YELLOW}Step 6: Setup Swap (2GB)${NC}"
if [ ! -f /swapfile ]; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    echo "vm.swappiness=10" >> /etc/sysctl.conf
    sysctl -p
    echo -e "${GREEN}Swap created.${NC}"
else
    echo "Swap file already exists."
fi

echo -e "\n${GREEN}===============================================${NC}"
echo -e "${GREEN}    Setup Complete!                            ${NC}"
echo -e "${GREEN}===============================================${NC}"
if [ "$BLOCK_ROOT_LOGIN" = true ]; then
    echo -e "${RED}IMPORTANT: Root login is disabled. Please test SSH login as '$ADMIN_USER' BEFORE disconnecting!${NC}"
fi
echo -e "You may need to logout and login again for group changes to take effect."
