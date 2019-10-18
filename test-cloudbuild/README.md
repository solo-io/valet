This is a work in progress, exploring initializing CI by invoking git clone instead of starting 
from an archive. This means we end up with git history in CI and can use `git describe --tags --always`
to come up with a meaningful version. 