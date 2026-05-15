#!/usr/bin/env bash
# Usage:
#   gen-skill-mounts.sh [file]        # read skills from file (one per line)
#   ls ~/.agents/skills | gen-skill-mounts.sh   # pipe in a list
#   gen-skill-mounts.sh skills.txt > skill-mounts.json

input="${1:--}"  # default to stdin

printf '// Bind-mount individual skills\n'
while IFS= read -r skill; do
  [[ -z "$skill" ]] && continue
  printf '    "source=${localEnv:HOME}/.agents/skills/%s,target=${containerWorkspaceFolder}/.claude/skills/%s,type=bind,consistency=cached,readonly",\n' \
    "$skill" "$skill"
done < "${input}"
