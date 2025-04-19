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
    local special="!@#%^&*_-"
    
    # Function to get a random character from a string using /dev/urandom
    function get_random_char() {
        local chars=$1
        local size=${#chars}
        local index=$(od -An -N2 -i /dev/urandom | tr -d ' ' | awk "{print \$1 % ${size}}")
        echo "${chars:${index}:1}"
    }
    
    # Initialize password with required characters
    local password=""
    password+="$(get_random_char "${lower}")"   # Ensure one lowercase
    password+="$(get_random_char "${upper}")"   # Ensure one uppercase
    password+="$(get_random_char "${digits}")"  # Ensure one digit
    password+="$(get_random_char "${special}")" # Ensure one special char
    
    # Add remaining characters
    local all_chars="${lower}${upper}${digits}${special}"
    for ((i=0; i<8; i++)); do
        password+="$(get_random_char "${all_chars}")"
    done
    
    # Shuffle the password using cryptographically secure randomness
    local shuffled=""
    local temp="$password"
    
    while [ ${#temp} -gt 0 ]; do
        local pos=$(od -An -N2 -i /dev/urandom | tr -d ' ' | awk "{print \$1 % ${#temp}}")
        shuffled+="${temp:$pos:1}"
        temp="${temp:0:$pos}${temp:$((pos+1))}"
    done
    
    echo "$shuffled"
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
## Generate a random password
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

## output crednetials inside the box
echo "Email: $EMAIL" >> /home/vagrant/credentials.txt
echo "Password: $PASSWORD" >> /home/vagrant/credentials.txt

