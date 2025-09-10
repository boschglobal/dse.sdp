
function dse-builder() {
    ( if test -f "$1"; then cd $(dirname "$1"); fi && docker run -it --user $(id -u):$(id -g) --rm -e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT -v $(pwd):/workdir $BUILDER_IMAGE "$@"; )
}
export -f dse-builder

function dse-simer() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --user $(id -u):$(id -g) --rm -v $(pwd):/sim -p 2159:2159 -p 6379:6379 $DSE_SIMER_IMAGE "$@"; )
}
export -f dse-simer

function dse-simer-host() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -it --user $(id -u):$(id -g) --rm --network=host -v $(pwd):/sim $DSE_SIMER_IMAGE "$@"; )
}
export -f dse-simer-host

function dse-report() {
    ( if test -d "$1"; then cd "$1" && shift; fi && docker run -t --user $(id -u):$(id -g) --rm -v $(pwd):/sim $DSE_REPORT_IMAGE /sim "$@"; )
}
export -f dse-report

alias dse-env='env | grep ^DSE | sort'
alias simer=dse-simer
alias h='history'
alias hi='history | grep $1'
