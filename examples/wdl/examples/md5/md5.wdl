version 1.0

workflow md5Workflow {
    
    input {
        File inputFile
        String knownMd5Sum
    }
    call md5 { input: inputFile=inputFile} 
    call checkMd5 { input: inputFile=md5.value,knownMd5Sum=knownMd5Sum }
    meta {
        author: "Brian O’Connor"
        email: "brian@somewhere.com"
        description: "a simple workflow that calculates an md5 checksum and then checks it"
    }

}

task md5 {
    input {
        File inputFile
    }

    command {
        /bin/my_md5sum ${inputFile}
    }

    output {
        File value = "md5sum.txt"
    }

    runtime {
    docker: "quay.io/briandoconnor/dockstore-tool-md5sum:1.0.4"
    cpu: 1
    memory: "512 MB"
    }

    parameter_meta {
        inputFile: "the file to create an md5 checksum for"
    }

}

task checkMd5 {
    input {
        File inputFile
        String knownMd5Sum
    }

    command {
        grep ${knownMd5Sum} ${inputFile} | wc -l > check_md5sum.report.txt
    }

    output {
        File value = "check_md5sum.report.txt"
    }

    runtime {
    docker: "quay.io/briandoconnor/dockstore-tool-md5sum:1.0.4"
    cpu: 1
    memory: "512 MB"
    }

    parameter_meta {
        inputFile: "the file to create an md5 checksum for"
        knownMd5Sum: "the known md5sum value to compare against"
    }

}
