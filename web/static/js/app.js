class VideoConverter {
    constructor() {
        // 自动检测API地址，支持不同部署环境
        this.API_BASE = window.location.protocol === 'file:' 
            ? 'http://localhost:8080/api/v1'  // 本地开发
            : window.location.hostname === '47.93.190.244'
                ? 'http://47.93.190.244:9002/api/v1'  // 生产环境IP访问
                : `${window.location.protocol}//${window.location.host}/api/v1`; // 其他环境
        this.currentTaskId = null;
        this.currentFile = null;
        this.progressInterval = null;
        this.currentUrl = null;
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.setupDragAndDrop();
        this.loadTasks();
    }
    
    setupEventListeners() {
        // 标签页切换
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.addEventListener('click', (e) => this.switchTab(e));
        });
        
        // 文件选择
        document.getElementById('select-file-btn').addEventListener('click', () => {
            document.getElementById('file-input').click();
        });
        
        document.getElementById('file-input').addEventListener('change', (e) => {
            this.handleFileSelect(e);
        });
        
        // URL提交
        document.getElementById('url-submit-btn').addEventListener('click', () => {
            this.handleUrlSubmit();
        });
        
        document.getElementById('video-url').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.handleUrlSubmit();
            }
        });
        
        // 文件移除
        document.getElementById('remove-file-btn').addEventListener('click', () => {
            this.removeFile();
        });
        
        // 转换开始
        document.getElementById('convert-btn').addEventListener('click', () => {
            this.startConversion();
        });
        
        // 取消转换
        document.getElementById('cancel-btn').addEventListener('click', () => {
            this.cancelConversion();
        });
        
        // 下载文件
        document.getElementById('download-btn').addEventListener('click', () => {
            this.downloadFile();
        });
        
        // 新转换
        document.getElementById('new-convert-btn').addEventListener('click', () => {
            this.resetToUpload();
        });
        
        // 刷新任务
        document.getElementById('refresh-tasks-btn').addEventListener('click', () => {
            this.loadTasks();
        });
    }
    
    setupDragAndDrop() {
        const uploadZone = document.getElementById('upload-zone');
        
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            uploadZone.addEventListener(eventName, this.preventDefaults, false);
        });
        
        ['dragenter', 'dragover'].forEach(eventName => {
            uploadZone.addEventListener(eventName, () => {
                uploadZone.classList.add('dragover');
            }, false);
        });
        
        ['dragleave', 'drop'].forEach(eventName => {
            uploadZone.addEventListener(eventName, () => {
                uploadZone.classList.remove('dragover');
            }, false);
        });
        
        uploadZone.addEventListener('drop', (e) => {
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                this.handleFile(files[0]);
            }
        }, false);
    }
    
    preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    switchTab(e) {
        const tabName = e.target.dataset.tab;
        
        // 更新导航按钮状态
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        e.target.classList.add('active');
        
        // 切换标签页内容
        document.querySelectorAll('.tab-content').forEach(tab => {
            tab.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');
        
        // 如果切换到任务列表，刷新数据
        if (tabName === 'tasks') {
            this.loadTasks();
        }
    }
    
    handleFileSelect(e) {
        const file = e.target.files[0];
        if (file) {
            this.handleFile(file);
        }
    }
    
    handleFile(file) {
        // 验证文件类型
        if (!file.type.startsWith('video/')) {
            this.showToast('请选择视频文件', 'error');
            return;
        }
        
        // 验证文件大小 (限制为500MB)
        const maxSize = 500 * 1024 * 1024;
        if (file.size > maxSize) {
            this.showToast('文件大小超过500MB限制', 'error');
            return;
        }
        
        this.currentFile = file;
        this.showFileInfo(file);
        this.showSettings();
    }
    
    showFileInfo(file) {
        const fileInfo = document.getElementById('file-info');
        const fileName = document.getElementById('file-name');
        const fileSize = document.getElementById('file-size');
        
        fileName.textContent = file.name;
        fileSize.textContent = this.formatFileSize(file.size);
        
        fileInfo.style.display = 'block';
    }
    
    removeFile() {
        this.currentFile = null;
        document.getElementById('file-info').style.display = 'none';
        document.getElementById('file-input').value = '';
        this.hideSettings();
    }
    
    async handleUrlSubmit() {
        const urlInput = document.getElementById('video-url');
        const url = urlInput.value.trim();
        
        if (!url) {
            this.showToast('请输入视频链接', 'error');
            return;
        }
        
        // 提取真实的视频链接
        const extractedUrl = this.extractVideoUrl(url);
        if (!extractedUrl) {
            this.showToast('未找到有效的视频链接', 'error');
            return;
        }
        
        // 显示提取到的链接（如果不同）
        if (extractedUrl !== url) {
            urlInput.value = extractedUrl;
            this.showToast(`已提取视频链接: ${extractedUrl}`, 'info');
        }
        
        if (!this.isValidUrl(extractedUrl)) {
            this.showToast('请输入有效的视频链接', 'error');
            return;
        }
        
        // 保存URL并显示设置
        this.currentUrl = extractedUrl;
        this.showSettings();
        this.showToast('请配置转换设置后开始转换', 'info');
    }
    
    // 新增：从分享文本中提取视频URL
    extractVideoUrl(text) {
        // 支持的抖音URL格式
        const douyinPatterns = [
            // 标准抖音链接
            /https?:\/\/v\.douyin\.com\/[A-Za-z0-9]+\/?/g,
            // 移动端链接
            /https?:\/\/www\.douyin\.com\/video\/\d+/g,
            // 其他可能的格式
            /https?:\/\/(?:www\.)?douyin\.com\/[^\\s]+/g
        ];
        
        // 支持的快手URL格式
        const kuaishouPatterns = [
            /https?:\/\/v\.kuaishou\.com\/[A-Za-z0-9]+\/?/g,
            /https?:\/\/www\.kuaishou\.com\/[^\\s]+/g
        ];
        
        // 合并所有模式
        const allPatterns = [...douyinPatterns, ...kuaishouPatterns];
        
        // 首先检查是否已经是有效的URL
        if (this.isValidUrl(text)) {
            return text;
        }
        
        // 尝试从文本中提取URL
        for (const pattern of allPatterns) {
            const matches = text.match(pattern);
            if (matches && matches.length > 0) {
                // 返回第一个匹配的URL
                let url = matches[0];
                
                // 清理URL末尾的标点符号
                url = url.replace(/[.,;!?。，；！？]*$/, '');
                
                // 确保URL以/结尾（对于v.douyin.com链接）
                if (url.includes('v.douyin.com') && !url.endsWith('/')) {
                    url += '/';
                }
                
                return url;
            }
        }
        
        return null;
    }
    
    async startConversion() {
        // 检查是否有文件或URL
        if (!this.currentFile && !this.currentUrl) {
            this.showToast('请先选择文件或输入视频链接', 'error');
            return;
        }
        
        // 立即显示进度界面
        this.showProgress();
        
        try {
            let response;
            
            if (this.currentUrl) {
                // URL转换
                response = await this.apiCall('/convert/url', 'POST', {
                    url: this.currentUrl,
                    title: document.getElementById('file-title').value || 'URL视频转换',
                    audio_codec: 'libmp3lame',
                    audio_bitrate: document.getElementById('bitrate-select').value,
                    sample_rate: document.getElementById('sample-rate-select').value
                });
                
                if (response.success) {
                    this.currentTaskId = response.data.task_id;
                    this.startProgressTracking();
                    this.showToast('URL转换任务已开始', 'success');
                } else {
                    this.showToast(response.message || '创建转换任务失败', 'error');
                    this.resetToUpload();
                }
            } else {
                // 文件转换
                // 1. 上传文件（带进度显示）
                const uploadResponse = await this.uploadFile();
                if (!uploadResponse.success) {
                    this.showToast(uploadResponse.message || '文件上传失败', 'error');
                    this.resetToUpload();
                    return;
                }
                
                this.currentTaskId = uploadResponse.data.task_id;
                
                // 2. 开始转换
                const convertResponse = await this.apiCall('/convert/file', 'POST', {
                    task_id: this.currentTaskId,
                    audio_codec: 'libmp3lame',
                    audio_bitrate: document.getElementById('bitrate-select').value,
                    sample_rate: document.getElementById('sample-rate-select').value
                });
                
                if (convertResponse.success) {
                    this.startProgressTracking();
                    this.showToast('转换任务已开始', 'success');
                } else {
                    this.showToast(convertResponse.message || '开始转换失败', 'error');
                    this.resetToUpload();
                }
            }
            
        } catch (error) {
            console.error('转换失败:', error);
            this.showToast('转换失败，请稍后重试', 'error');
            this.resetToUpload();
        }
    }
    
    async uploadFile() {
        const formData = new FormData();
        formData.append('file', this.currentFile);
        formData.append('title', document.getElementById('file-title').value || this.currentFile.name);
        formData.append('description', '通过前端界面上传的视频文件');
        
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            
            // 监听上传进度
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    this.updateUploadProgress(percentComplete);
                }
            });
            
            // 监听上传完成
            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    try {
                        const response = JSON.parse(xhr.responseText);
                        resolve(response);
                    } catch (error) {
                        reject(new Error('解析响应失败'));
                    }
                } else {
                    reject(new Error(`上传失败: ${xhr.status}`));
                }
            });
            
            // 监听错误
            xhr.addEventListener('error', () => {
                reject(new Error('网络错误'));
            });
            
            // 发送请求
            xhr.open('POST', `${this.API_BASE}/upload`);
            xhr.send(formData);
        });
    }
    
    updateUploadProgress(percent) {
        const statusText = document.getElementById('status-text');
        const progressPercent = document.getElementById('progress-percent');
        const progressFill = document.getElementById('progress-fill');
        const estimatedTime = document.getElementById('estimated-time');
        
        statusText.textContent = '上传中...';
        progressPercent.textContent = `${Math.round(percent)}%`;
        progressFill.style.width = `${percent}%`;
        
        if (percent < 100) {
            estimatedTime.textContent = `上传进度：${Math.round(percent)}%`;
        } else {
            statusText.textContent = '上传完成，准备转换...';
            estimatedTime.textContent = '上传完成，正在初始化转换任务...';
        }
        
        // 添加调试信息
        console.log(`上传进度: ${Math.round(percent)}%`);
    }
    
    startProgressTracking() {
        this.progressInterval = setInterval(async () => {
            try {
                const response = await this.apiCall(`/tasks/${this.currentTaskId}/status`);
                if (response.success) {
                    this.updateProgress(response.data);
                    
                    if (response.data.status === 'completed') {
                        this.stopProgressTracking();
                        this.showDownload();
                    } else if (response.data.status === 'failed') {
                        this.stopProgressTracking();
                        this.showToast('转换失败: ' + (response.data.error_message || '未知错误'), 'error');
                        this.resetToUpload();
                    }
                }
            } catch (error) {
                console.error('获取进度失败:', error);
            }
        }, 2000); // 每2秒检查一次
    }
    
    stopProgressTracking() {
        if (this.progressInterval) {
            clearInterval(this.progressInterval);
            this.progressInterval = null;
        }
    }
    
    updateProgress(data) {
        const statusText = document.getElementById('status-text');
        const progressPercent = document.getElementById('progress-percent');
        const progressFill = document.getElementById('progress-fill');
        
        const statusMap = {
            'queued': '排队中...',
            'processing': '转换中...',
            'completed': '转换完成',
            'failed': '转换失败',
            'canceled': '已取消'
        };
        
        statusText.textContent = statusMap[data.status] || data.status;
        progressPercent.textContent = `${Math.round(data.progress)}%`;
        progressFill.style.width = `${data.progress}%`;
    }
    
    async cancelConversion() {
        if (!this.currentTaskId) return;
        
        try {
            const response = await this.apiCall(`/tasks/${this.currentTaskId}/cancel`, 'POST');
            if (response.success) {
                this.stopProgressTracking();
                this.showToast('转换已取消', 'warning');
                this.resetToUpload();
            }
        } catch (error) {
            this.showToast('取消失败', 'error');
        }
    }
    
    async downloadFile() {
        if (!this.currentTaskId) return;
        
        try {
            // 显示下载进度
            this.showDownloadProgress();
            
            const response = await fetch(`${this.API_BASE}/download/${this.currentTaskId}`);
            
            if (!response.ok) {
                throw new Error(`下载失败: ${response.status}`);
            }
            
            const contentLength = response.headers.get('Content-Length');
            const total = contentLength ? parseInt(contentLength, 10) : 0;
            
            const reader = response.body.getReader();
            const chunks = [];
            let loaded = 0;
            
            while (true) {
                const { done, value } = await reader.read();
                
                if (done) break;
                
                chunks.push(value);
                loaded += value.length;
                
                if (total > 0) {
                    const percent = (loaded / total) * 100;
                    this.updateDownloadProgress(percent);
                }
            }
            
            // 合并数据
            const blob = new Blob(chunks, { type: 'audio/mpeg' });
            
            // 创建下载链接
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            const fileName = this.currentFile ? 
                this.currentFile.name.replace(/\.[^/.]+$/, ".mp3") : 
                "转换后的音频.mp3";
            
            a.href = url;
            a.download = fileName;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            // 下载完成
            this.hideDownloadProgress();
            this.showToast('文件下载完成', 'success');
            
        } catch (error) {
            console.error('下载失败:', error);
            this.hideDownloadProgress();
            this.showToast('下载失败，请稍后重试', 'error');
        }
    }
    
    showDownloadProgress() {
        // 创建下载进度显示元素
        const downloadSection = document.getElementById('download-section');
        const progressHtml = `
            <div class="download-progress" id="download-progress" style="margin-bottom: 1rem;">
                <div class="progress-status">
                    <span class="status-text">下载中...</span>
                    <span class="progress-percent" id="download-progress-percent">0%</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" id="download-progress-fill" style="width: 0%;"></div>
                </div>
            </div>
        `;
        
        // 在下载按钮前插入进度条
        const downloadActions = downloadSection.querySelector('.download-actions');
        downloadActions.insertAdjacentHTML('beforebegin', progressHtml);
        
        // 禁用下载按钮
        document.getElementById('download-btn').disabled = true;
    }
    
    updateDownloadProgress(percent) {
        const progressPercent = document.getElementById('download-progress-percent');
        const progressFill = document.getElementById('download-progress-fill');
        
        if (progressPercent && progressFill) {
            progressPercent.textContent = `${Math.round(percent)}%`;
            progressFill.style.width = `${percent}%`;
        }
    }
    
    hideDownloadProgress() {
        const downloadProgress = document.getElementById('download-progress');
        if (downloadProgress) {
            downloadProgress.remove();
        }
        
        // 重新启用下载按钮
        document.getElementById('download-btn').disabled = false;
    }
    
    async loadTasks() {
        const container = document.getElementById('tasks-container');
        const loading = document.getElementById('loading-tasks');
        
        loading.style.display = 'block';
        
        try {
            const response = await this.apiCall('/tasks');
            if (response.success && response.data.items) {
                this.renderTasks(response.data.items);
            } else {
                container.innerHTML = '<p class="no-tasks">暂无转换任务</p>';
            }
        } catch (error) {
            container.innerHTML = '<p class="error-tasks">加载任务失败，请稍后重试</p>';
        }
        
        loading.style.display = 'none';
    }
    
    renderTasks(tasks) {
        const container = document.getElementById('tasks-container');
        
        if (tasks.length === 0) {
            container.innerHTML = '<p class="no-tasks">暂无转换任务</p>';
            return;
        }
        
        const tasksHtml = tasks.map(task => `
            <div class="task-item">
                <div class="task-header">
                    <div>
                        <h4 class="task-title">${task.title}</h4>
                    </div>
                    <span class="task-status ${task.status}">${this.getStatusText(task.status)}</span>
                </div>
                <div class="task-meta">
                    <span>创建时间: ${this.formatDate(task.created_at)}</span>
                    <span>进度: ${Math.round(task.progress)}%</span>
                </div>
                ${this.getTaskActions(task)}
            </div>
        `).join('');
        
        container.innerHTML = tasksHtml;
        
        // 绑定任务操作事件
        this.bindTaskActions();
    }
    
    getStatusText(status) {
        const statusMap = {
            'queued': '排队中',
            'processing': '处理中',
            'completed': '已完成',
            'failed': '失败',
            'canceled': '已取消'
        };
        return statusMap[status] || status;
    }
    
    getTaskActions(task) {
        if (task.status === 'completed') {
            return `
                <div class="task-actions">
                    <button class="task-btn download" onclick="app.downloadTask('${task.id}')">
                        下载MP3
                    </button>
                </div>
            `;
        } else if (task.status === 'queued' || task.status === 'processing') {
            return `
                <div class="task-actions">
                    <button class="task-btn cancel" onclick="app.cancelTask('${task.id}')">
                        取消任务
                    </button>
                </div>
            `;
        }
        return '';
    }
    
    bindTaskActions() {
        // 任务操作事件已通过onclick属性绑定
    }
    
    async downloadTask(taskId) {
        try {
            // 显示下载进度Toast
            this.showToast('开始下载...', 'info');
            
            const response = await fetch(`${this.API_BASE}/download/${taskId}`);
            
            if (!response.ok) {
                throw new Error(`下载失败: ${response.status}`);
            }
            
            const contentLength = response.headers.get('Content-Length');
            const contentDisposition = response.headers.get('Content-Disposition');
            
            // 从Content-Disposition头提取文件名
            let fileName = '转换后的音频.mp3';
            if (contentDisposition) {
                const fileNameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
                if (fileNameMatch && fileNameMatch[1]) {
                    fileName = fileNameMatch[1].replace(/['"]/g, '');
                }
            }
            
            const total = contentLength ? parseInt(contentLength, 10) : 0;
            const reader = response.body.getReader();
            const chunks = [];
            let loaded = 0;
            
            while (true) {
                const { done, value } = await reader.read();
                
                if (done) break;
                
                chunks.push(value);
                loaded += value.length;
                
                if (total > 0) {
                    const percent = (loaded / total) * 100;
                    // 可以在这里更新进度，但为了简化UI，我们只在控制台显示
                    if (Math.round(percent) % 10 === 0) {
                        console.log(`下载进度: ${Math.round(percent)}%`);
                    }
                }
            }
            
            // 合并数据并下载
            const blob = new Blob(chunks, { type: 'audio/mpeg' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            
            a.href = url;
            a.download = fileName;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            this.showToast('文件下载完成', 'success');
            
        } catch (error) {
            console.error('下载失败:', error);
            this.showToast('下载失败，请稍后重试', 'error');
        }
    }
    
    async cancelTask(taskId) {
        try {
            const response = await this.apiCall(`/tasks/${taskId}/cancel`, 'POST');
            if (response.success) {
                this.showToast('任务已取消', 'warning');
                this.loadTasks();
            }
        } catch (error) {
            this.showToast('取消失败', 'error');
        }
    }
    
    showSettings() {
        document.getElementById('settings-section').style.display = 'block';
    }
    
    hideSettings() {
        document.getElementById('settings-section').style.display = 'none';
    }
    
    showProgress() {
        document.getElementById('settings-section').style.display = 'none';
        document.getElementById('progress-section').style.display = 'block';
        document.getElementById('download-section').style.display = 'none';
    }
    
    async showDownload() {
        document.getElementById('progress-section').style.display = 'none';
        document.getElementById('download-section').style.display = 'block';
        
        // 获取任务详细信息
        try {
            const response = await this.apiCall(`/tasks/${this.currentTaskId}`);
            if (response.success && response.data) {
                const task = response.data;
                
                // 更新文件名
                const fileName = task.original_name ? 
                    task.original_name.replace(/\.[^/.]+$/, ".mp3") : 
                    "转换后的音频.mp3";
                
                document.getElementById('download-file-name').textContent = fileName;
                
                // 更新文件大小
                const fileSizeElement = document.getElementById('download-file-size');
                if (task.output_size && task.output_size > 0) {
                    fileSizeElement.textContent = `大小：${this.formatFileSize(task.output_size)}`;
                } else {
                    // 如果任务信息中没有文件大小，尝试通过HEAD请求获取
                    this.getFileSize();
                }
            }
        } catch (error) {
            console.error('获取任务信息失败:', error);
            // 如果获取任务信息失败，使用默认值
            const fileName = this.currentFile ? 
                this.currentFile.name.replace(/\.[^/.]+$/, ".mp3") : 
                "转换后的音频.mp3";
            
            document.getElementById('download-file-name').textContent = fileName;
            this.getFileSize();
        }
    }
    
    async getFileSize() {
        try {
            const response = await fetch(`${this.API_BASE}/download/${this.currentTaskId}`, {
                method: 'HEAD'
            });
            
            if (response.ok) {
                const contentLength = response.headers.get('Content-Length');
                const fileSizeElement = document.getElementById('download-file-size');
                
                if (contentLength) {
                    const size = parseInt(contentLength, 10);
                    fileSizeElement.textContent = `大小：${this.formatFileSize(size)}`;
                } else {
                    fileSizeElement.textContent = '大小：未知';
                }
            }
        } catch (error) {
            console.error('获取文件大小失败:', error);
            document.getElementById('download-file-size').textContent = '大小：未知';
        }
    }
    
    resetToUpload() {
        this.currentTaskId = null;
        this.currentFile = null;
        this.currentUrl = null;
        this.stopProgressTracking();
        
        document.getElementById('file-input').value = '';
        document.getElementById('file-info').style.display = 'none';
        document.getElementById('settings-section').style.display = 'none';
        document.getElementById('progress-section').style.display = 'none';
        document.getElementById('download-section').style.display = 'none';
        document.getElementById('video-url').value = '';
        document.getElementById('file-title').value = '';
    }
    
    async apiCall(endpoint, method = 'GET', data = null) {
        const url = `${this.API_BASE}${endpoint}`;
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        if (data && method !== 'GET') {
            options.body = JSON.stringify(data);
        }
        
        const response = await fetch(url, options);
        return await response.json();
    }
    
    showToast(message, type = 'info') {
        const toast = document.getElementById('toast');
        const toastIcon = document.getElementById('toast-icon');
        const toastMessage = document.getElementById('toast-message');
        
        // 设置图标
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️'
        };
        
        toast.className = `toast ${type}`;
        toastIcon.textContent = icons[type] || icons.info;
        toastMessage.textContent = message;
        
        // 显示消息
        toast.classList.add('show');
        
        // 3秒后自动隐藏
        setTimeout(() => {
            toast.classList.remove('show');
        }, 3000);
    }
    
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    isValidUrl(string) {
        try {
            const url = new URL(string);
            return url.protocol === 'http:' || url.protocol === 'https:';
        } catch (_) {
            return false;
        }
    }
}

// 初始化应用
const app = new VideoConverter(); 