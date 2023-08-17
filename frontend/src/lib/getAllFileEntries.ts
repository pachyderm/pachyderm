const getAllFileEntries = async (
  dataTransferItemList: DataTransferItemList,
) => {
  const fileEntries: FileSystemFileEntry[] = [];
  // traverse entire directory/file structure
  const queue: ReturnType<DataTransferItem['webkitGetAsEntry']>[] = [];
  for (let i = 0; i < dataTransferItemList.length; i++) {
    queue.push(dataTransferItemList[i].webkitGetAsEntry());
  }
  while (queue.length > 0) {
    const entry = queue.shift();
    if (entry) {
      if (entry.isFile) {
        fileEntries.push(entry as FileSystemFileEntry);
      } else if (entry.isDirectory) {
        queue.push(
          ...(await readAllDirectoryEntries(
            (entry as FileSystemDirectoryEntry).createReader(),
          )),
        );
      }
    }
  }
  return fileEntries;
};

// Get all the entries (files or sub-directories) in a directory
// by calling readEntries until it returns empty array
async function readAllDirectoryEntries(
  directoryReader: FileSystemDirectoryReader,
) {
  const entries = [];
  let readEntries = await readEntriesPromise(directoryReader);

  while ((readEntries || []).length > 0) {
    entries.push(...(readEntries || []));
    readEntries = await readEntriesPromise(directoryReader);
  }
  return entries;
}

// Wrap readEntries in a promise to make working with readEntries easier
async function readEntriesPromise(directoryReader: FileSystemDirectoryReader) {
  return await new Promise<FileSystemEntry[]>((resolve) => {
    directoryReader.readEntries(resolve, (err) => err.message);
  });
}

export default getAllFileEntries;
