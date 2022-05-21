#!/usr/bin/env bash
APP_ENV=$1
case $APP_ENV in 
    'dev')
        ;;
    'prod')
        ;;
    *)
        APP_ENV='prod'
        ;;
esac
echo "environment = $APP_ENV"

#程序名称
APP=nec
PROJECT_HOME=$(cd $(dirname $0) && cd .. && pwd -P)
PKG_HOME="$PROJECT_HOME/outputs"

cd $PROJECT_HOME

function doBuild() {
    GO=$GOROOT/bin/go
    $GO version

    pwd
    echo -ne "-> building $1 \t ... "
    make #>/dev/null
    if [ $? -eq 0 ]; then
        echo 'done'
    else
        exit 1
    fi
}

rm -rf $PKG_HOME 2>/dev/null
mkdir -p $PKG_HOME/{bin,conf,log,run}
cp scripts/conf/$APP_ENV.ini $PKG_HOME/conf/$APP.ini
cp scripts/bin/run_${APP_ENV}.sh $PKG_HOME/bin/run.sh

echo 'building started'
doBuild
echo 'building finished'
