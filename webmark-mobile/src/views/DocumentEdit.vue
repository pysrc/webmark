<script setup>
    import { ref, onMounted, reactive } from 'vue'
    import { computed } from 'vue'
    import { marked } from 'marked';
    import { useRouter, useRoute } from 'vue-router'
    import Cookies from 'js-cookie'
    import CryptoJS from 'crypto-js';

    const route = useRoute();
    const router = useRouter();
    const mdname = route.params.mdname;
    const groupname = route.params.groupname;

    const renderer = new marked.Renderer();
    renderer.link = ({href, title, text}) => {
        if (!href.startsWith('http')) {
            return `<a href="/wmapi/markdown/${groupname}/${href.replace(/^\/+/, '')}" title="${title || ''}">${text}</a>`;
        }
        return `<a href="${href}" title="${title || ''}">${text}</a>`;
    };
    renderer.image = ({href, _, text}) => {
        if (!href.startsWith('http')) {
            return `<img src="/wmapi/markdown/${groupname}/${href.replace(/^\/+/, '')}" alt="${text}" loading="lazy">`
        }
        return `<img src="${href}" alt="${text}" loading="lazy">`
    };
    marked.setOptions({ renderer });


    const activeTab = ref('edit');
    const editForm = reactive({
        title: mdname,
        content: ''
    });
    
    // 密码相关状态
    const password = ref('');
    
    // 预览内容
    const previewContent = computed(() => {
        return marked.parse(editForm.content || '');
    });

    onMounted(() => {
        fetch(`/wmapi/markdown/${groupname}/${mdname}.md?_t=${Date.now()}`, {
            method: 'GET',
            headers: {
                'Cache-Control': 'no-cache'
            }
        })
            .then(response => response.text())
            .then(d => {
                editForm.content = d;
            });
    });
    
    const goBack = () => {
        router.go(-1);
    };

    const saveMarkdown = () => {
        fetch(`/wmapi/update-markdown/${groupname}/${mdname}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/text'
            },
            body: editForm.content
        })
            .then(response => response.json())
            .then(d => {
                if (d.ok) {
                    // 通知保存成功
                    alert('保存成功');
                    router.go(-1);
                } else {
                    alert("保存失败");
                }
            });
    };
    
    const saveDocument = () => {
        saveMarkdown();
    };
    
    const handleLogout = () => {
        fetch('/wmapi/logout', {
            method: 'GET'
        })
        .then(response => response.json())
        .then(d => {
            if (d.ok) {
                Cookies.remove("session_id");
                Cookies.remove("username");
                localStorage.removeItem('isLoggedIn');
                router.push('/');
            } else {
                alert(d.msg);
            }
        });
    };

    // 加密函数
    const encryptContent = () => {
        if (!password.value) {
            alert('请输入密码');
            return;
        }
        
        try {
            var _enmd = CryptoJS.AES.encrypt(editForm.content, password.value).toString();
            editForm.content = _enmd;
        } catch (error) {
            alert('加密失败: ' + error.message);
        }
    };

    // 解密函数
    const decryptContent = () => {
        if (!password.value) {
            alert('请输入密码');
            return;
        }
        
        try {
            var _demd = CryptoJS.AES.decrypt(editForm.content, password.value).toString(CryptoJS.enc.Utf8);
            editForm.content = _demd;
        } catch (error) {
            alert('解密失败: ' + error.message);
        }
    };
</script>
<template>
    <div class="system-container">
        <div class="header">
        <div class="header-title">编辑文档</div>
        <button class="logout-btn" @click="handleLogout">退出登录</button>
        </div>
        <div class="content">
        <button class="back-btn" @click="goBack">
            <span>←</span> 返回文档
        </button>
        
        <!-- 密码和加密解密功能 -->
        <div class="security-section">
            <div class="password-input-group">
                <input 
                    type="password" 
                    class="password-input" 
                    v-model="password"
                    placeholder="输入加密/解密密码"
                />
                <div class="encryption-buttons">
                    <button 
                        class="encrypt-btn" 
                        @click="encryptContent"
                    >
                        密
                    </button>
                    <button 
                        class="decrypt-btn" 
                        @click="decryptContent"
                    >
                        解
                    </button>
                </div>
            </div>
        </div>
        
        <div class="edit-container">
            <div class="edit-tabs">
            <button 
                class="edit-tab" 
                :class="{ active: activeTab === 'edit' }"
                @click="activeTab = 'edit'"
            >
                编辑
            </button>
            <button 
                class="edit-tab" 
                :class="{ active: activeTab === 'preview' }"
                @click="activeTab = 'preview'"
            >
                预览
            </button>
            </div>
            
            <div class="edit-panel" v-if="activeTab === 'edit'">
                <div class="doc-title-section">
                    <h1 class="doc-detail-title">{{ editForm.title }}</h1>
                </div>
                <textarea 
                    class="edit-textarea" 
                    v-model="editForm.content"
                    placeholder="请输入Markdown内容..."
                ></textarea>
            </div>
            
            <div class="preview-panel" v-if="activeTab === 'preview'">
            <h1>{{ editForm.title }}</h1>
            <div v-html="previewContent"></div>
            </div>
            
            <div class="edit-actions">
            <button class="cancel-btn" @click="goBack">取消</button>
            <button class="save-btn" @click="saveDocument">保存</button>
            </div>
        </div>
        </div>
    </div>
</template>
<style scoped>
    /* 系统页面样式 */
    .system-container {
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    .header {
      background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
      color: white;
      padding: 15px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    }
    
    .header-title {
      font-size: 20px;
      font-weight: 600;
    }

    .logout-btn {
      background: rgba(255, 255, 255, 0.2);
      color: white;
      border: none;
      padding: 8px 15px;
      border-radius: 20px;
      font-size: 14px;
      cursor: pointer;
    }
    .content {
      flex: 1;
      padding: 20px;
      overflow-y: auto;
    }
    .back-btn {
      background: #f0f0f0;
      border: none;
      padding: 10px 15px;
      border-radius: 8px;
      margin-bottom: 20px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 5px;
    }

    /* 安全功能样式 */
    .security-section {
        background: #f8f9fa;
        border: 1px solid #e9ecef;
        border-radius: 8px;
        padding: 15px;
        margin-bottom: 20px;
    }

    .password-input-group {
        display: flex;
        gap: 10px;
        align-items: center;
        margin-bottom: 10px;
    }

    .password-input {
        flex: 1;
        padding: 10px 15px;
        border: 1px solid #ddd;
        border-radius: 6px;
        font-size: 14px;
    }

    .password-input:focus {
        border-color: #6a11cb;
        outline: none;
    }

    .encryption-buttons {
        display: flex;
        gap: 8px;
    }

    .encrypt-btn, .decrypt-btn {
        padding: 10px 15px;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        cursor: pointer;
        transition: all 0.2s;
    }

    .encrypt-btn {
        background: #28a745;
        color: white;
    }

    .encrypt-btn:hover:not(:disabled) {
        background: #218838;
    }

    .decrypt-btn {
        background: #17a2b8;
        color: white;
    }

    .decrypt-btn:hover:not(:disabled) {
        background: #138496;
    }

    .encrypt-btn:disabled, .decrypt-btn:disabled {
        background: #6c757d;
        cursor: not-allowed;
        opacity: 0.6;
    }
    
    /* 编辑模式样式 */
    .edit-container {
      display: flex;
      flex-direction: column;
      height: calc(100vh - 180px);
    }
    
    .edit-tabs {
      display: flex;
      border-bottom: 1px solid #ddd;
      margin-bottom: 15px;
    }
    
    .edit-tab {
      padding: 10px 20px;
      background: none;
      border: none;
      cursor: pointer;
      border-bottom: 2px solid transparent;
      transition: all 0.2s;
    }
    
    .edit-tab.active {
      border-bottom-color: #6a11cb;
      color: #6a11cb;
      font-weight: 600;
    }
    
    .edit-panel {
      flex: 1;
      display: flex;
      flex-direction: column;
    }
    
    .edit-textarea {
      flex: 1;
      width: 100%;
      padding: 15px;
      border: 1px solid #ddd;
      border-radius: 8px;
      font-family: 'Courier New', monospace;
      font-size: 14px;
      line-height: 1.5;
      resize: none;
      margin-bottom: 15px;
    }
    
    .edit-textarea:focus {
      border-color: #6a11cb;
      outline: none;
    }
    
    .preview-panel {
      flex: 1;
      padding: 15px;
      border: 1px solid #ddd;
      border-radius: 8px;
      overflow-y: auto;
      background: white;
    }
    
    .edit-actions {
      display: flex;
      gap: 10px;
      margin-top: 15px;
    }
    
    .save-btn {
      background: #6a11cb;
      color: white;
      border: none;
      padding: 10px 20px;
      border-radius: 6px;
      font-size: 14px;
      cursor: pointer;
      flex: 1;
    }
    
    .save-btn:hover {
      background: #5a0db5;
    }
    
    .cancel-btn {
      background: #f0f0f0;
      border: none;
      padding: 10px 20px;
      border-radius: 6px;
      font-size: 14px;
      cursor: pointer;
      flex: 1;
    }
    
    .cancel-btn:hover {
      background: #e0e0e0;
    }

    .doc-title-section {
      flex: 1;
    }
    
    .doc-detail-title {
      font-size: 22px;
      font-weight: 600;
      margin-bottom: 10px;
    }
</style>