# Export Your Data with `pachctl`
 
To export your data with pachctl:

1. List the files in the given directory:

      ```shell
      pachctl list file <repo>@<branch>:<dir>
      ```

      **Example:**
      ```shell
      pachctl list myrepo@master:labresults
      ```

      **System Response:**
      ```shell
      NAME                                                    TYPE SIZE
      /labresults/T1606331395-LIPID-PATID2-CLIA24D9871327.txt file 101B
      /labresults/T1606707557-LIPID-PATID1-CLIA24D9871327.txt file 101B
      /labresults/T1606707579-LIPID-PATID3-CLIA24D9871327.txt file 100B
      /labresults/T1606707597-LIPID-PATID4-CLIA24D9871327.txt file 101B
      ```

1. Get the contents of a specific file:

      ```shell
      pachctl get file <repo>@<branch>:<path/to/file>
      ```

      **Example:**
      ```shell
      pachctl get file myrepo@master:/labresults/T1606331395-LIPID-PATID2-CLIA24D9871327.txt
      ```

      **System Response:**
      ```shell
      PID|PATID2
      ORC|ORD777889
      OBX|1|NM|2093-3^Cholesterol|212|mg/dL
      OBX|2|NM|2571-8^Triglyceride|110|mg/dL
      ```

!!! Note
      You can view the parent, grandparent, and any previous
      commit by using the caret (`^`) symbol followed by a number that
      corresponds to an ancestor in sequence:

      * View a parent commit
         ```shell
         pachctl list commit <repo>@<branch-or-commit>^:<path/to/file>
         ```

         ```shell
         pachctl get file <repo>@<branch-or-commit>^:<path/to/file>
         ```

      * View an `<n>` parent of a commit
         ```shell
         pachctl list commit <repo>@<branch-or-commit>^<n>:<path/to/file>
         ```

         ```shell
         pachctl get file <repo>@<branch-or-commit>^<n>:<path/to/file>
         ```

         **Example:**
         ```shell
         pachctl get file datas@master^4:user_data.csv
         ```

         If the file does not exist in that revision, Pachyderm displays an error message.

