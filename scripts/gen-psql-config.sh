
#!/bin/sh
cd $1
PES_DIR=astrolabe_conf/pes
mkdir -p $PES_DIR
PSQL_FILE=$PES_DIR/psql.pe.json
echo "{" > $PSQL_FILE
echo "	\"snapshotsDir\":\"/astrolabe-repo\"" >> $PSQL_FILE
echo "}" >> $PSQL_FILE

