#!/bin/bash
set +e

if [ -z $1 ] || ([ $1 != 'create' ] && [ $1 != 'delete' ])
then
    echo "Invalid option, valid values => [ create, delete ]"
    exit 1
fi

function create_cluster {
    eksctl create cluster --name=$1 --version=$2 --region=us-east-1
}

function delete_cluster {
    eksctl delete cluster --name=$1
}

echo "Selected Option: "$1

if [ $1 = 'create' ]
then
    create_cluster $2 $3
fi


if [ $1 = 'delete' ]
then
    delete_cluster $2
fi
