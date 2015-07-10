package pfs

type apiServer struct{}

func newAPIServer() *apiServer {
	return &apiServer{}
}

func (a *apiServer) GetFile(getFileRequest *GetFileRequest, apiGetFileServer Api_GetFileServer) error {
	return nil
}
