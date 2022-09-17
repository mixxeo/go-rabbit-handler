#!/usr/bin/env bash

sudo apt install -y rabbitmq-server
echo "AMQP_URI=\"amqp://skku:1234@localhost:5672/%2f\"" > .env

# install rabbitmqadmin
wget http://localhost:15672/cli/rabbitmqadmin
chmod +x rabbitmqadmin
sudo mv rabbitmqadmin /etc/rabbitmq

# add user and permission
# rabbitmqctl add_user skku 1234
# rabbitmqctl set_user_tags skku administrator
# rabbitmqctl set_permissions -p / skku ".*" ".*" ".*"

# Make an Exchange
rabbitmqadmin -u skku -p 1234 declare exchange name=judger-exchange type=direct

# Make queues
rabbitmqadmin -u skku -p 1234 declare queue name=submission-queue durable=true
rabbitmqadmin -u skku -p 1234 declare queue name=result-queue durable=true

# Make bindings
rabbitmqadmin -u skku -p 1234 declare binding source=judger-exchange\
                                destination_type=queue destination=submission-queue routing_key=submission
rabbitmqadmin -u skku -p 1234 declare binding source=judger-exchange\
                                destination_type=queue destination=result-queue routing_key=result


