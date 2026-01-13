package handler

//func Upload(c *gin.Context) {
//	file, err := c.FormFile("file")
//	if err != nil {
//		c.JSON(400, gin.H{"error": "no file uploaded"})
//		return
//	}
//
//	// 生成文件名
//	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
//
//	// 打开上传的文件
//	src, err := file.Open()
//	if err != nil {
//		c.JSON(500, gin.H{"error": "failed to open uploaded file"})
//		return
//	}
//	defer src.Close()
//
//	// 保存文件到本地（临时）
//	localPath, err := storageImpl.SaveUploadedFile(src, filename)
//	if err != nil {
//		c.JSON(500, gin.H{"error": "failed to save file"})
//		return
//	}
//
//	// 如果是七牛云存储，上传到云端并删除本地文件
//	var fileURL string
//	if cfg.Storage.Type == "qiniu" && cfg.Qiniu.Enabled {
//		// 生成七牛云存储key（使用uploads目录前缀）
//		key := fmt.Sprintf("uploads/%s", filename)
//		cloudURL, err := storageImpl.UploadFile(localPath, key)
//		if err != nil {
//			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to upload to cloud: %v", err)})
//			return
//		}
//		fileURL = cloudURL
//
//		// 删除本地临时文件
//		if err := storageImpl.DeleteLocalFile(localPath); err != nil {
//			log.Printf("Warning: failed to delete local file %s: %v", localPath, err)
//		}
//	} else {
//		fileURL = fmt.Sprintf("/api/uploads/%s", filename)
//	}
//
//	c.JSON(200, gin.H{
//		"filename": filename,
//		"path":     localPath,
//		"url":      fileURL,
//	})
//}
