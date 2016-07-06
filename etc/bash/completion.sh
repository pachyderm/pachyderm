# bash completion for ./pachctl                            -*- shell-script -*-

__debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__my_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__handle_reply()
{
    __debug "${FUNCNAME[0]}"
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%%=*}"
                __index_of_word "${flag}" "${flags_with_completion[@]}"
                if [[ ${index} -ge 0 ]]; then
                    COMPREPLY=()
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zfs completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions=("${must_have_one_noun[@]}")
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
    fi

    __ltrim_colon_completions "$cur"
}

# The arguments should be in the form "ext1|ext2|extn"
__handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1
}

__handle_flag()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    if [ -n "${flagvalue}" ] ; then
        flaghash[${flagname}]=${flagvalue}
    elif [ -n "${words[ $((c+1)) ]}" ] ; then
        flaghash[${flagname}]=${words[ $((c+1)) ]}
    else
        flaghash[${flagname}]="true" # pad "true" for bool flag
    fi

    # skip the argument to a two word flag
    if __contains_word "${words[c]}" "${two_word_flags[@]}"; then
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__handle_noun()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__handle_command()
{
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_$(basename "${words[c]//:/__}")"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F $next_command >/dev/null && $next_command
}

__handle_word()
{
    if [[ $c -ge $cword ]]; then
        __handle_reply
        return
    fi
    __debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __handle_flag
    elif __contains_word "${words[c]}" "${commands[@]}"; then
        __handle_command
    elif [[ $c -eq 0 ]] && __contains_word "$(basename "${words[c]}")" "${commands[@]}"; then
        __handle_command
    else
        __handle_noun
    fi
    __handle_word
}

_./pachctl_repo()
{
    last_command="./pachctl_repo"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_create-repo()
{
    last_command="./pachctl_create-repo"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_inspect-repo()
{
    last_command="./pachctl_inspect-repo"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-repo()
{
    last_command="./pachctl_list-repo"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--provenance=")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--provenance=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_delete-repo()
{
    last_command="./pachctl_delete-repo"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force")
    flags+=("-f")
    local_nonpersistent_flags+=("--force")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_commit()
{
    last_command="./pachctl_commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_start-commit()
{
    last_command="./pachctl_start-commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--parent=")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--parent=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_finish-commit()
{
    last_command="./pachctl_finish-commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cancel")
    flags+=("-c")
    local_nonpersistent_flags+=("--cancel")
    flags+=("--debug")
    flags+=("-d")
    local_nonpersistent_flags+=("--debug")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_inspect-commit()
{
    last_command="./pachctl_inspect-commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-commit()
{
    last_command="./pachctl_list-commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    flags+=("--block")
    flags+=("-b")
    local_nonpersistent_flags+=("--block")
    flags+=("--provenance=")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--provenance=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_flush-commit()
{
    last_command="./pachctl_flush-commit"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--repos=")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--repos=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-branch()
{
    last_command="./pachctl_list-branch"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_file()
{
    last_command="./pachctl_file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_put-file()
{
    last_command="./pachctl_put-file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_get-file()
{
    last_command="./pachctl_get-file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--block-modulus=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--block-modulus=")
    flags+=("--block-shard=")
    two_word_flags+=("-b")
    local_nonpersistent_flags+=("--block-shard=")
    flags+=("--file-modulus=")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--file-modulus=")
    flags+=("--file-shard=")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--file-shard=")
    flags+=("--from=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--from=")
    flags+=("--unsafe")
    local_nonpersistent_flags+=("--unsafe")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_inspect-file()
{
    last_command="./pachctl_inspect-file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--block-modulus=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--block-modulus=")
    flags+=("--block-shard=")
    two_word_flags+=("-b")
    local_nonpersistent_flags+=("--block-shard=")
    flags+=("--file-modulus=")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--file-modulus=")
    flags+=("--file-shard=")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--file-shard=")
    flags+=("--unsafe")
    local_nonpersistent_flags+=("--unsafe")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-file()
{
    last_command="./pachctl_list-file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--block-modulus=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--block-modulus=")
    flags+=("--block-shard=")
    two_word_flags+=("-b")
    local_nonpersistent_flags+=("--block-shard=")
    flags+=("--file-modulus=")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--file-modulus=")
    flags+=("--file-shard=")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--file-shard=")
    flags+=("--from=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--from=")
    flags+=("--unsafe")
    local_nonpersistent_flags+=("--unsafe")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_delete-file()
{
    last_command="./pachctl_delete-file"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_mount()
{
    last_command="./pachctl_mount"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--block-modulus=")
    two_word_flags+=("-n")
    local_nonpersistent_flags+=("--block-modulus=")
    flags+=("--block-shard=")
    two_word_flags+=("-b")
    local_nonpersistent_flags+=("--block-shard=")
    flags+=("--file-modulus=")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--file-modulus=")
    flags+=("--file-shard=")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--file-shard=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_job()
{
    last_command="./pachctl_job"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_create-job()
{
    last_command="./pachctl_create-job"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_inspect-job()
{
    last_command="./pachctl_inspect-job"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--block")
    flags+=("-b")
    local_nonpersistent_flags+=("--block")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_get-logs()
{
    last_command="./pachctl_get-logs"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-job()
{
    last_command="./pachctl_list-job"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--pipeline=")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--pipeline=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_pipeline()
{
    last_command="./pachctl_pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_create-pipeline()
{
    last_command="./pachctl_create-pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_inspect-pipeline()
{
    last_command="./pachctl_inspect-pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_list-pipeline()
{
    last_command="./pachctl_list-pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_delete-pipeline()
{
    last_command="./pachctl_delete-pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_run-pipeline()
{
    last_command="./pachctl_run-pipeline"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_version()
{
    last_command="./pachctl_version"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl_delete-all()
{
    last_command="./pachctl_delete-all"
    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_./pachctl()
{
    last_command="./pachctl"
    commands=()
    commands+=("repo")
    commands+=("create-repo")
    commands+=("inspect-repo")
    commands+=("list-repo")
    commands+=("delete-repo")
    commands+=("commit")
    commands+=("start-commit")
    commands+=("finish-commit")
    commands+=("inspect-commit")
    commands+=("list-commit")
    commands+=("flush-commit")
    commands+=("list-branch")
    commands+=("file")
    commands+=("put-file")
    commands+=("get-file")
    commands+=("inspect-file")
    commands+=("list-file")
    commands+=("delete-file")
    commands+=("mount")
    commands+=("job")
    commands+=("create-job")
    commands+=("inspect-job")
    commands+=("get-logs")
    commands+=("list-job")
    commands+=("pipeline")
    commands+=("create-pipeline")
    commands+=("inspect-pipeline")
    commands+=("list-pipeline")
    commands+=("delete-pipeline")
    commands+=("run-pipeline")
    commands+=("version")
    commands+=("delete-all")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_./pachctl()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __my_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("./pachctl")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local nouns=()

    __handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_./pachctl ./pachctl
else
    complete -o default -o nospace -F __start_./pachctl ./pachctl
fi

# ex: ts=4 sw=4 et filetype=sh
