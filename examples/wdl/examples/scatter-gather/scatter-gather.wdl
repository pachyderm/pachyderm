version 1.0

workflow ga4ghMd5 {

    input {
        Array[File] inputFiles
    }

    scatter (oneFile in inputFiles) {
        call md5 { input: in=oneFile }
    }
    call md5ofmd5s { input: inputFiles=md5.value }
}

task md5 {
    input {
        File in
    }

    command {
        /bin/my_md5sum ${in}
    }

    output {
        File value = "md5sum.txt"
    }

    runtime {
        docker: "quay.io/briandoconnor/dockstore-tool-md5sum:1.0.4"
        cpu: 1
        memory: "512 MB"
    }
}

task md5ofmd5s {
    input{
        Array[File] inputFiles
    }

    command {
        md5sum ${sep="" inputFiles} > md5sum.report.txt
    }

    output {
        File value = "md5sum.report.txt"
    }

    runtime {
        docker: "quay.io/briandoconnor/dockstore-tool-md5sum:1.0.4"
        cpu: 1
        memory: "512 MB"
    }
}