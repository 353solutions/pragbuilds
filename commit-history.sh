#!/usr/bin/bash
# Commit notifications.json to git if it has changed.

file_name="notifications.json"

if [ -z "$(git status -s ${file_name})" ]; then
    echo "${file_name}: no change"
    exit
fi

git config user.name "GitHub Actions Bot"
git config user.email "<>"

git --no-pager diff "${file_name}"
git commit -m "notifications in run ${GITHUB_RUN_NUMBER:-<unknown>}" "${file_name}"
git push
