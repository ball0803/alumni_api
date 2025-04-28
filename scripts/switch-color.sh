#!/bin/bash
CONFIG_FILE="/etc/nginx/nginx.conf"

# Determine current active color
if grep -q "alumni-api-blue" $CONFIG_FILE; then
  NEW_COLOR="green"
  OLD_COLOR="blue"
else
  NEW_COLOR="blue"
  OLD_COLOR="green"
fi

# Update nginx config
sed -i "s/alumni-api-$OLD_COLOR/alumni-api-$NEW_COLOR/" $CONFIG_FILE

# Test and reload
nginx -t && nginx -s reload
echo "Switched from $OLD_COLOR to $NEW_COLOR"
