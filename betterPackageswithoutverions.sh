export MAIN="$(/bin/ls index/*/*.json   | xargs cat | jq 'select(.versions[0]  ==  "main")    | select ((.versions  | length) == 1) .name ' | wc -l)"
export MASTER="$(/bin/ls index/*/*.json | xargs cat | jq 'select(.versions[0]  ==  "master")  | select ((.versions  | length) == 1) .name ' | wc -l)"

echo "main: $MAIN"
echo "master: $MASTER"

echo $(($MAIN + $MASTER))
