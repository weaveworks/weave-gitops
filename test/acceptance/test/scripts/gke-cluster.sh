#!/bin/bash
set +e

if [ -z $1 ] || ([ $1 != 'create' ] && [ $1 != 'delete' ])
then
    echo "Invalid option, valid values => [ create, delete ]"
    exit 1
fi

function create_cluster {
    gcloud container clusters create $1 --cluster-version=$2
}

function delete_cluster {
    gcloud container clusters delete $1 --quiet
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
