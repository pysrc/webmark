<script setup>
    import { ref, onMounted } from 'vue'
    import { computed } from 'vue'
    import { marked } from 'marked';
    import { useRouter, useRoute } from 'vue-router'
    import Cookies from 'js-cookie'

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


    const route = useRoute();
    const router = useRouter();
    
    const mdname = route.params.mdname;
    const groupname = route.params.groupname;
    let document = ref({ title: mdname, content: '' });
    
    // 使用marked渲染Markdown内容
    const renderedContent = computed(() => {
        return marked.parse(document.value.content || '');
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
                document.value = {
                    title: mdname,
                    content: d,
                };
            });
    });
    
    const goBack = () => {
        router.go(-1);
    };
    
    const editDocument = () => {
        router.push(`/edit/${groupname}/${mdname}`);
    };

    const deleteDocument = () => {
        const confirmed = confirm(`
        ⚠️ 删除确认
        您即将删除文档: "${document.value.title}"
        此操作不可恢复，确定要继续吗？
        `.trim());
        
        if (confirmed) {
            // 删除逻辑...
            fetch(`/wmapi/del-markdown/${groupname}/${mdname}`, {
                method: 'DELETE'
            })
            .then(response => response.text())
            .then(d => {
                // 使用更友好的提示
                alert('✅ 文档已成功删除');
                router.push(`/group/${groupname}`);
            });
        }
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
</script>
<template>
    <div class="system-container">
        <div class="header">
        <div class="header-title">文档详情</div>
        <button class="logout-btn" @click="handleLogout">退出登录</button>
        </div>
        <div class="content">
        <button class="back-btn" @click="goBack">
            <span>←</span> 返回文档列表
        </button>
        <div class="doc-detail">
            <div class="doc-header">
            <div class="doc-title-section">
                <h1 class="doc-detail-title">{{ document.title }}</h1>
            </div>
            <div class="doc-actions">
                <button class="action-btn edit" @click="editDocument">
                <span>✏️</span> 编辑
                </button>
                <button class="action-btn delete" @click="deleteDocument">
                <span>🗑️</span> 删除
                </button>
            </div>
            </div>
            <div class="doc-content" v-html="renderedContent"></div>
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
    /* 文档详情样式 */
    .doc-detail {
      background: white;
      border-radius: 12px;
      padding: 20px;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);
    }
    
    .doc-header {
      margin-bottom: 20px;
      padding-bottom: 15px;
      border-bottom: 1px solid #f0f0f0;
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
    }
    
    .doc-title-section {
      flex: 1;
    }
    
    .doc-detail-title {
      font-size: 22px;
      font-weight: 600;
      margin-bottom: 10px;
    }
    
    .doc-meta {
      display: flex;
      gap: 15px;
      font-size: 14px;
      color: #777;
    }
    
    .doc-actions {
      display: flex;
      gap: 10px;
    }
    
    .action-btn {
      background: #f0f0f0;
      border: none;
      padding: 8px 12px;
      border-radius: 6px;
      font-size: 14px;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 5px;
      transition: background-color 0.2s;
    }
    
    .action-btn:hover {
      background: #e0e0e0;
    }
    
    .action-btn.edit {
      background: #6a11cb;
      color: white;
    }
    
    .action-btn.edit:hover {
      background: #5a0db5;
    }
    
    .doc-content {
      line-height: 1.8;
    }
    
    .doc-content h1, .doc-content h2, .doc-content h3 {
      margin-top: 1.5em;
      margin-bottom: 0.5em;
    }
    
    .doc-content p {
      margin-bottom: 1em;
    }
    
    .doc-content ul, .doc-content ol {
      margin-bottom: 1em;
      padding-left: 2em;
    }
    
    .doc-content blockquote {
      border-left: 4px solid #6a11cb;
      padding-left: 1em;
      margin: 1em 0;
      color: #666;
    }
    
    .doc-content code {
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 4px;
      font-family: 'Courier New', monospace;
    }
    
    .doc-content pre {
      background: #f8f8f8;
      padding: 1em;
      border-radius: 8px;
      overflow-x: auto;
      margin: 1em 0;
    }
    
    .doc-content pre code {
      background: none;
      padding: 0;
    }
</style>