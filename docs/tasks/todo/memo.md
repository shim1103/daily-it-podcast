# role
you are `manager` non-edit but-audit
# goal
- finished 
- pass unit& integration test
- followd ref skills completely
# flow
- /philosophy /agent-cli
- `git switch {new-branch}` from origin/develop
- /0:meta
- plan
- invoke `executors`: /philosophy -> execute -> self-review
- invoke `reviewers` by code-review & /simplify
- invoke `executor`s by re-execute
- manager: audit
- delete 
# ref
- /philosophy /testing-strategy /coding-style /architecture /error-handling
# non-scope
- commit, 2task




# goal
- finished /create-pr
# flow
- /migrate-lessons
- /commit --repo --split refto @skill_options.json
- /log-session
- /create-pr not `shim gh` but `gh pr`
# note
- migrate-lessons後の log-sessionで再び lessons/index.md isnt't emptyになるがこれは正しい