setup:
	mkdir -p ./labresults 
	printf "PID|PATID1\nORC|ORD777888\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID1-CLIA24D9871327.txt
	printf "PID|PATID2\nORC|ORD777889\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID2-CLIA24D9871327.txt
	printf "PID|PATID3\nORC|ORD777890\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID3-CLIA24D9871327.txt
	printf "PID|PATID4\nORC|ORD777891\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID4-CLIA24D9871327.txt
	printf "PID|PATID1\nORC|ORD777892\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID1-CLIA24D9871328.txt
	printf "PID|PATID3\nORC|ORD777893\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID3-CLIA24D9871328.txt

deploy:
	pachctl create repo labresults
	pachctl put file -r labresults@master:/ -f labresults
	pachctl list file labresults@master
	pachctl create pipeline -f group.json


