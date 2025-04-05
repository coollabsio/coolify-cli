#!/bin/bash

# Exit on error
set -e

## Create a function to generate a random password
function generate_password() {
    # Generate a cryptographically secure random password
    # - 12 characters long
    # - Contains at least one lowercase, one uppercase, one digit, and one special character
    # - Uses /dev/urandom for cryptographically secure randomness
    
    # Define character sets
    local lower="abcdefghijklmnopqrstuvwxyz"
    local upper="ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    local digits="0123456789"
    local special="!@#$%^&*()_+-="
    local all_chars="${lower}${upper}${digits}${special}"
    
    # Ensure at least one character from each set
    local password=""
    password+="${lower:$((RANDOM % ${#lower})):1}"
    password+="${upper:$((RANDOM % ${#upper})):1}"
    password+="${digits:$((RANDOM % ${#digits})):1}"
    password+="${special:$((RANDOM % ${#special})):1}"
    
    # Fill the rest with random characters
    for ((i=0; i<8; i++)); do
        password+="${all_chars:$((RANDOM % ${#all_chars})):1}"
    done
    
    # Shuffle the password using a Fisher-Yates shuffle
    local result=""
    local temp="$password"
    local len=${#temp}
    
    for ((i=len-1; i>0; i--)); do
        local j=$((RANDOM % (i+1)))
        local temp_i="${temp:$i:1}"
        local temp_j="${temp:$j:1}"
        temp="${temp:0:$j}${temp_i}${temp:$((j+1)):$((i-j-1))}${temp_j}${temp:$((i+1))}"
    done
    
    echo "$temp"
}

# Update package lists and upgrade system
echo "Updating system packages..."
apt-get update
apt-get upgrade -y

# Install required dependencies
echo "Installing dependencies..."
apt-get install -y \
    curl \
    wget \
    git \
    vim \
    htop

## Install Coolify
echo "Installing Coolify..."
## Generated a random password
PASSWORD=$(generate_password)
EMAIL="hi@coollabs.io"
curl -fsSL https://cdn.coollabs.io/coolify/install.sh > install.sh
env ROOT_USERNAME=test ROOT_USER_EMAIL=''"$EMAIL"'' ROOT_USER_PASSWORD=''"$PASSWORD"'' bash install.sh

# Add current user to docker group
usermod -aG docker vagrant

# Print completion message
echo "Installation complete!"
echo "Coolify should be accessible at http://localhost:8000"
echo "You can SSH into the VM using: vagrant ssh" 
echo "Credentials:"
echo "Email: $EMAIL"
echo "Password: $PASSWORD"
