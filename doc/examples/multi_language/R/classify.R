library('e1071');
library('SparseM');
library('tm');

# LOAD DATA FROM CSV
Sample_data <- read.csv("Path of .csv file");

# CREATE DATA FRAME OF 750 TRAINING JOURNALS ARTICLES AND 250
# TEST ARTICLES INCLUDING 'Abstract'(Column 1) AND 'Journal_group' (Column 2)
traindata <- as.data.frame(Sample_data[1:750,c(1,2)]);
testdata <- as.data.frame(Sample_data[751:1000,c(1,2)]);

# SEPARATE TEXT VECTOR TO CREATE Source(),
# Corpus() CONSTRUCTOR FOR DOCUMENT TERM
# MATRIX TAKES Source()
trainvector <- as.vector(traindata$Abstract);
testvector <- as.vector(testdata$Abstract);

# CREATE SOURCE FOR VECTORS
trainsource <- VectorSource(trainvector);
testsource <- VectorSource(testvector);

# CREATE CORPUS FOR DATA
traincorpus <- Corpus(trainsource);
testcorpus <- Corpus(testsource);

# PERFORMING THE VARIOUS TRANSFORMATION on "traincorpus" and "testcorpus" DATASETS #SUCH AS TRIM WHITESPACE, REMOVE PUNCTUATION, REMOVE STOPWORDS.
traincorpus <- tm_map(traincorpus,stripWhitespace);
traincorpus <- tm_map(traincorpus,tolower);
traincorpus <- tm_map(traincorpus, removeWords,stopwords("english"));
traincorpus<- tm_map(traincorpus,removePunctuation);
traincorpus <- tm_map(traincorpus, PlainTextDocument);
testcorpus <- tm_map(testcorpus,stripWhitespace);
testcorpus <- tm_map(testcorpus,tolower);
testcorpus <- tm_map(testcorpus, removeWords,stopwords("english"));
testcorpus<- tm_map(testcorpus,removePunctuation);
testcorpus <- tm_map(testcorpus, PlainTextDocument);

# CREATE TERM DOCUMENT MATRIX
trainmatrix <- t(TermDocumentMatrix(traincorpus));
testmatrix <- t(TermDocumentMatrix(testcorpus));

# TRAIN NAIVE BAYES MODEL USING trainmatrix DATA AND traindate$Journal_group CLASS VECTOR
model <- naiveBayes(as.matrix(trainmatrix),as.factor(traindata$Journal_group));

# PREDICTION
results <- predict(model,as.matrix(testmatrix));
