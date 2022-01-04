#!/bin/bash

# Pulled and modified from this link:
# https://gist.github.com/devster/b91b97ebbca4db4d02b84337b2a3d933

# Script to simplify the release flow.
# 1) Fetch the current release version
# 2) Increase the version (major, minor, patch)
# 3) Add a new git tag
# 4) Push the new tag

# Parse command line options.
while getopts ":Mmpcdr" Option
do
  case $Option in
    M  ) major=true;;
    m  ) minor=true;;
    p  ) patch=true;;
    c  ) candidate=true;;
    d  ) dry=true;;
    r  ) release=true;;
  esac
done

shift $(($OPTIND - 1))

# Display usage
if [ -z $major ] && [ -z $minor ] && [ -z $patch ] && [ -z $dry ] && [ -z $release ] [ -z $candidate];
then
  echo "usage: $(basename $0) [Mmp]"
  echo ""
  echo "  -d Dry run"
  echo "  -M for a major release candidate"
  echo "  -m for a minor release candidate"
  echo "  -p for a patch release candidate"
  echo "  -c for incrementing the release candidate number"
  echo "  -r for a full release"
  echo ""
  echo " Example: release -p"
  echo " means create a patch release candidate"
  exit 1
fi

# 1) Fetch the current release version

echo "Fetch tags"
git fetch --prune origin +refs/tags/*:refs/tags/*

version=$(git describe --abbrev=0 --tags)
echo "Current version: $version"

version=${version:1} # Remove the v in the tag v0.37.10 for example

# Build array from version string.
a=( ${version//./ } )
# Get rc number from version string
rc=${version#*rc} 
a[2]=${a[2]/-rc*} # Clean up -rc on patch number

# 2) Set version number

# Increment version/rc numbers as requested.
if [ ! -z $major ]
then
  ((a[0]++))
  a[1]=0
  a[2]=0
  rc=0
fi

if [ ! -z $minor ]
then
  ((a[1]++))
  a[2]=0
  rc=0
fi

if [ ! -z $patch ]
then
  ((a[2]++))
  rc=0
fi

if [ ! -z $candidate ]
then
  ((rc++)) 
fi

if [ ! -z $release ]
then 
  next_version="${a[0]}.${a[1]}.${a[2]}"
else 
  next_version="${a[0]}.${a[1]}.${a[2]}-rc${rc}"
fi

if [ "$version" == "$next_version" ]
then
    echo "version did not change from current version: v$version to next version: v$next_version"
    exit 1
fi

# If its a dry run, just display the new release version number
if [ ! -z $dry ]
then
  echo "Next version: v$next_version"
else
  # If a command fails, exit the script
  set -e

  # If it's not a dry run, let's go!
  # 3) Add git tag
  echo "Add git tag v$next_version with title: v$next_version"
  git tag -a "v$next_version" -m "v$next_version"

  # 4) Push the new tag

  echo "Push the tag"
  git push --tags origin main

  echo -e "Release done: $next_version"
fi