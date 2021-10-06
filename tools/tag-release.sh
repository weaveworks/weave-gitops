#!/bin/bash

# Pulled and modified from this link:
# https://gist.github.com/devster/b91b97ebbca4db4d02b84337b2a3d933

# Script to simplify the release flow.
# 1) Fetch the current release version
# 2) Increase the version (major, minor, patch)
# 3) Add a new git tag
# 4) Push the tag

# Parse command line options.
while getopts ":Mmpdr" Option
do
  case $Option in
    M ) major=true;;
    m ) minor=true;;
    p ) patch=true;;
    d ) dry=true;;
    r ) release=true;;
  esac
done

shift $(($OPTIND - 1))

# Display usage
if [ -z $major ] && [ -z $minor ] && [ -z $patch ] && [ -z $dry ] && [ -z $release ];
then
  echo "usage: $(basename $0) [Mmp]"
  echo ""
  echo "  -d Dry run"
  echo "  -M for a major release candidate"
  echo "  -m for a minor release candidate"
  echo "  -p for a patch release candidate"
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
version=${version:1} # Remove the v in the tag v0.37.10 for example

echo "Current version: $version"

# Build array from version string.

a=( ${version//./ } )


# 2) Increase version number if release candidate and if not create release

if [ ! -z $release ] 
then
  if [[ $version == *"rc"* ]]; then # If the last version wasnt a release candidate exit
    patch=( ${a[2]//-/ } )
    next_version="${a[0]}.${a[1]}.${patch[0]}"
    echo $next_version
  else
    echo "previous release was not a release candidate"
    exit 1
  fi
else
# Increment version numbers as requested.
  if [ ! -z $major ]
  then
    ((a[0]++))
    a[1]=0
    a[2]=0
  fi

  if [ ! -z $minor ]
  then
    ((a[1]++))
    a[2]=0
  fi

  if [ ! -z $patch ]
  then
    ((a[2]++))
  fi

  next_version="${a[0]}.${a[1]}.${a[2]}-rc"
fi

msg="$1"

branch=$(git branch | sed -n -e 's/^\* \(.*\)/\1/p')

# If its a dry run, just display the new release version number
if [ ! -z $dry ]
then
  echo "Next version: v$next_version"
else
  # If a command fails, exit the script
  set -e

  # Push main
  git push origin $branch

  # If it's not a dry run, let's go!
  # 3) Add git tag
  echo "Add git tag v$next_version with title: v$next_version"
  git tag -a "v$next_version" -m "v$next_version""

  # 4) Push the new tag

  echo "Push the tag"
  git push --tags origin $branch

  echo -e "\e[32mRelease done: $next_version\e[0m"
fi