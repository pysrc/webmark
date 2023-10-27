const _low_html = `
<div id="loweditor-toolbar" style="margin-top: 10px;display: none;">
<button id="loweditor-ap-strong" class="loweditor-tool-item"><b>B</b></button>
<button id="loweditor-ap-em" class="loweditor-tool-item"><em>I</em></button>
<button id="loweditor-ap-table" class="loweditor-tool-item"><b>插入表格&gt;</b></button>
<input id="loweditor-table-rows" class="loweditor-tool-input-item" placeholder="行">
<input id="loweditor-table-cols" class="loweditor-tool-input-item" placeholder="列">
<button id="loweditor-insert-newline" class="loweditor-tool-item"><b>插入新行</b></button>
</div>
<div id="loweditor-editor" class="loweditor-editor"></div>

<ul id="loweditor-custom-menu" style="position: fixed;">
    <li id="loweditor-rinsert">右边插入1列</li>
    <li id="loweditor-linsert">左边插入1列</li>
    <li id="loweditor-uinsert">上边插入1行</li>
    <li id="loweditor-dinsert">下边插入1行</li>
    <li id="loweditor-rdelete">删除行</li>
    <li id="loweditor-cdelete">删除列</li>
</ul>
<progress id="loweditor-progress" value="0" max="100" style="position: absolute; left: 0; right: 0; top: 50%; bottom: 50%; margin: auto;display: none;"></progress>
`;
const $$ = document.getElementById.bind(document);
const $c = document.createElement.bind(document);

function LowEditor(containerid, options) {
    $$(containerid).innerHTML = _low_html;
    const editor = $$("loweditor-editor");
    // 表格的右键菜单
    var custom_menu = $$("loweditor-custom-menu");
    // 进度条
    var pbar = $$("loweditor-progress");
    document.addEventListener("click", function(event){
        custom_menu.style.display = "none";
    });
    var menuEvent = function(event) {
        event.preventDefault();
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        originNode = anchorNode;
        custom_menu.style.display = "block";
        //获取鼠标视口位置
        custom_menu.style.top = event.clientY + "px";
        custom_menu.style.left = event.clientX + "px";
    }
    var enableMenu = function() {
        var tbodys = editor.querySelectorAll('tbody');
        for (let index = 0; index < tbodys.length; index++) {
            var tbody = tbodys[index];
            tbody.addEventListener("contextmenu", menuEvent);
        }
    }
    var disableMenu = function() {
        var tbodys = editor.querySelectorAll('tbody');
        for (let index = 0; index < tbodys.length; index++) {
            var tbody = tbodys[index];
            tbody.removeEventListener("contextmenu", menuEvent);
        }
    }
    var setEditable = function (enable) {
        editor.setAttribute("contenteditable", enable);
        if(enable) {
            $$("loweditor-toolbar").style.display=null;
            // 表格允许编辑
            enableMenu();
            // 禁止代码直接编辑
            // disablePre();
        } else {
            $$("loweditor-toolbar").style.display='none';
            // 表格不许编辑
            disableMenu();
        }
    }
    if(options.editable) {
        setEditable(true);
    }
    // 事件发生时的node
    var originNode = null;

    function insertDom(dom) {
        const selection = window.getSelection();
        var node = selection.anchorNode;
        if (node == null) {
            return;
        }
        var range = selection.getRangeAt(0);
        if (range.startOffset == range.endOffset) {
            // 没有选中则整行
            var text = node.textContent;
            range = new Range();
            range.setStart(node, 0);
            range.setEnd(node, text.length);
        }
        var ext = range.extractContents();
        dom.appendChild(ext);
        range.insertNode(dom);
        // 重新设置光标位置
        selection.removeAllRanges();
        let nrange = new Range();
        nrange.setStart(dom, 1);
        nrange.setEnd(dom, 1);
        window.getSelection().addRange(nrange);
    }
    function _insertImage(dom) {
        const selection = window.getSelection();
        const range = selection.getRangeAt(0);
        var ext = range.extractContents();
        dom.appendChild(ext);
        range.insertNode(dom);
        range.setStartAfter(dom);
    }
    function createTable(rows, cols) {
        var selection = window.getSelection();
        if (!selection.anchorNode.isContentEditable) {
            alert("请点击插入位置！");
            return;
        }
        var range = selection.getRangeAt(0);
        var ext = range.extractContents();

        var table = $c('table');
        table.style.border = "1px solid black";
        table.style.borderCollapse = "collapse";
        var tbody = $c('tbody');
        var thead = $c('thead');
        var tr = $c('tr');
        for (var j = 0; j < cols; j++) {
            var cell = $c('th');
            cell.innerHTML = `title${j + 1}`;
            cell.style.border = "1px solid black";
            tr.appendChild(cell);
        }
        thead.appendChild(tr);
        table.append(thead);
        for (var i = 0; i < rows; i++) {
            var row = $c('tr');
            for (var j = 0; j < cols; j++) {
                var cell = $c('td');
                cell.innerHTML = '<br>';
                cell.style.border = "1px solid black";
                row.appendChild(cell);
            }
            tbody.appendChild(row);
        }
        table.appendChild(tbody);
        table.appendChild(ext);
        range.insertNode(table);
        selection.removeAllRanges();
        enableMenu();
    }

    function _rlinsert(call) {
        var ptr = originNode.parentNode.parentNode;
        var td = originNode.parentNode;
        if(originNode.tagName&&originNode.tagName.toLowerCase() == "td") {
            ptr = originNode.parentNode;
            td = originNode;
        }
        var cur = 0;
        var tds = ptr.getElementsByTagName("td");
        for (let index = 0; index < tds.length; index++) {
            const e = tds[index];
            if(e == td) {
                cur = index;
                break;
            }
        }
        // 插入头
        var ths = ptr.parentNode.parentNode.getElementsByTagName('th');
        var th = $c('th');
        th.innerText = 'insert';
        th.style.border = "1px solid black";
        call(ths[cur], th);
        // 插入表格
        var tbody = ptr.parentNode;
        var trs = tbody.getElementsByTagName('tr');
        for (let index = 0; index < trs.length; index++) {
            const tr = trs[index];
            var td = $c('td');
            td.innerHTML = '<br>';
            td.style.border = "1px solid black";
            call(tr.children[cur], td);
        }
    }
    $$("loweditor-rinsert").addEventListener("click", function(){
        // 右边插入
        _rlinsert((a, b)=>{
            a.after(b);
        });
    });
    $$("loweditor-linsert").addEventListener("click", function(){
        // 左边插入
        _rlinsert((a, b)=>{
            a.before(b);
        });
    });
    function _udinsert(call) {
        var ptr = originNode.parentNode.parentNode;
        if(originNode.tagName&&originNode.tagName.toLowerCase() == "td") {
            ptr = originNode.parentNode;
        }
        var tr=ptr.cloneNode(true);
        var tds = tds=tr.getElementsByTagName('td');
        for (let index = 0; index < tds.length; index++) {
            const td = tds[index];
            td.innerHTML = '<br>';
        }
        call(ptr, tr);
    }
    $$("loweditor-uinsert").addEventListener("click", function(){
        // 上边插入
        _udinsert((ptr, tr) => {
            ptr.before(tr);
        });
    });
    $$("loweditor-dinsert").addEventListener("click", function(){
        // 下边插入
        _udinsert((ptr, tr) => {
            ptr.after(tr);
        });
    });
    $$("loweditor-rdelete").addEventListener("click", function(){
        // 删除行
        var ptr = originNode.parentNode.parentNode;
        if(originNode.tagName&&originNode.tagName.toLowerCase() == "td") {
            ptr = originNode.parentNode;
        }
        ptr.remove();
    });
    $$("loweditor-cdelete").addEventListener("click", function(){
        // 删除列
        var ptr = originNode.parentNode.parentNode;
        var td = originNode.parentNode;
        if(originNode.tagName&&originNode.tagName.toLowerCase() == "td") {
            ptr = originNode.parentNode;
            td = originNode;
        }
        var cur = 0;
        var tds = ptr.getElementsByTagName("td");
        for (let index = 0; index < tds.length; index++) {
            const e = tds[index];
            if(e == td) {
                cur = index;
                break;
            }
        }
        // 删除头
        var ths = ptr.parentNode.parentNode.getElementsByTagName('th');
        ths[cur].remove();
        // 删除表格
        var tbody = ptr.parentNode;
        var trs = tbody.getElementsByTagName('tr');
        for (let index = 0; index < trs.length; index++) {
            trs[index].children[cur].remove();
        }
    });
    function handleHeader() {
        // 获取光标
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        var anchorOffset = selection.anchorOffset;
        // 获取光标所在元素的内容，判断生成几级标题
        var text = anchorNode.textContent;
        var level = text.match(/#+/)[0].length;
        var newLineNode = $c(`h${level}`);
        newLineNode.textContent = text.replace(/#+\u00a0|#+ /, '');
        anchorNode.after(newLineNode);
        anchorNode.remove();
        // 重新设置光标位置
        selection.removeAllRanges();
        let nrange = new Range();
        nrange.setStart(newLineNode, 1);
        nrange.setEnd(newLineNode, 1);
        window.getSelection().addRange(nrange);
    }
    function handleList() {
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        var text = anchorNode.textContent;
        text = text.replace(/\*[ |\u00a0]/, '');
        var parent = anchorNode.parentNode;
        var ul = $c('ul');
        var li = $c('li');
        li.innerText = text;
        ul.appendChild(li);
        anchorNode.after(ul);
        anchorNode.remove();
        // 重新设置光标位置
        selection.removeAllRanges();
        let nrange = new Range();
        nrange.setStart(ul, 1);
        nrange.setEnd(ul, 1);
        window.getSelection().addRange(nrange);
    }
    function handleSortedList() {
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        var text = anchorNode.textContent;
        var start = text.match(/\d+/)[0];
        text = text.replace(/\d+\.[ |\u00a0]/, '');
        var ol = $c('ol');
        ol.setAttribute("start", start);
        var li = $c('li');
        li.innerText = text;
        ol.appendChild(li);
        anchorNode.after(ol);
        anchorNode.remove();
        // 重新设置光标位置
        selection.removeAllRanges();
        let nrange = new Range();
        nrange.setStart(ol, 1);
        nrange.setEnd(ol, 1);
        window.getSelection().addRange(nrange);
    }
    function handleRef() {
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        var text = anchorNode.textContent;
        text = text.replace(/>[ |\u00a0]/, '');
        var parent = anchorNode.parentNode;
        var blockquote = $c('blockquote');
        var p = $c('p');
        p.innerText = text + "\n";
        blockquote.appendChild(p);
        anchorNode.after(blockquote);
        anchorNode.remove();
        blockquote.after($c("br"));
        // 重新设置光标位置
        selection.removeAllRanges();
        let nrange = new Range();
        nrange.setStart(blockquote, 1);
        nrange.setEnd(blockquote, 1);
        window.getSelection().addRange(nrange);
    }
    function handleTable() {
        var rows = $$("loweditor-table-rows").value;
        var cols = $$("loweditor-table-cols").value;
        if (!rows) {
            alert("请输入插入表格行数！");
            return;
        }
        if (!cols) {
            alert("请输入插入表格列数！");
            return;
        }
        createTable(rows, cols);
    }
    function applyStyle(style) {
        const element = $c(style);
        insertDom(element);
    }
    function insertImage(url) {
        const element = $c('img');
        element.src = url;
        _insertImage(element);
    }
    function insertUrl(url, filename) {
        const element = $c('a');
        element.setAttribute('href', url);
        element.setAttribute('target', '_blank');
        element.innerText = filename;
        _insertImage(element);
    }
    // 文件上传
    function upload(files) {
        pbar.style.display = "block";
        // 创建FormData对象，用于将文件上传到服务器
        var formData = new FormData();
        var image_names = [];
        var file_names = [];
        // 将拖拽的文件添加到FormData对象中
        for (var i = 0; i < files.length; i++) {
            var name = `${Date.now()}_${files[i].name}`;
            formData.append('file', files[i], name);
            if (files[i].type.indexOf('image') !== -1) {
                image_names.push(name);
            } else {
                file_names.push(name);
            }
        }
        // 创建XMLHttpRequest对象，用于发送请求
        var xhr = new XMLHttpRequest();
        // 监听上传进度
        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                var percent = (e.loaded / e.total) * 100;
                console.log('上传进度：' + percent + '%');
                pbar.value = percent;
            }
        });
        // 监听上传完成事件
        xhr.addEventListener('load', (e) => {
            pbar.style.display = "none";
            console.log('上传完成');
            for (var i = 0; i < image_names.length; i++) {
                insertImage(`${options.file.prefix()}${image_names[i]}`);
            }
            for (var i = 0; i < file_names.length; i++) {
                insertUrl(`${options.file.prefix()}${file_names[i]}`, file_names[i]);
            }
        });
        // 监听上传出错事件
        xhr.addEventListener('error', (e) => {
            console.log('上传出错');
        });
        // 监听上传取消事件
        xhr.addEventListener('abort', (e) => {
            console.log('上传取消');
        });
        // 发送请求
        xhr.open('POST', `${options.file.upload()}`);
        xhr.send(formData);
    };
    if (options.file) {
        editor.addEventListener('drop', (e) => {
            e.preventDefault();
            e.stopPropagation();
            var files = e.dataTransfer.files;
            upload(files);
        });
        // 粘贴文件、截图等
        editor.addEventListener('paste', (e) => {
            const clipboardData = e.clipboardData;
            if (clipboardData.types.includes('Files')) {
                e.preventDefault();
                e.stopPropagation();
                var files = clipboardData.files;
                upload(files);
            }
        });
    }
    editor.addEventListener('input', (e) => {
        var selection = window.getSelection();
        var anchorNode = selection.anchorNode;
        var p = anchorNode.parentNode; // 父节点如果不是编辑器，则不管
        if(p && (p.tagName.toLowerCase() === 'div' || p.tagName.toLowerCase() === 'p') && p.parentNode && p.parentNode.id === 'loweditor-editor') {
            var text = anchorNode.textContent;
            if (text.match(/#+[ |\u00a0]./)) {
                handleHeader();
            } else if (text.match(/\*[ |\u00a0]/)) {
                handleList();
            } else if (text.match(/\d+\.[ |\u00a0]/)) {
                handleSortedList();
            } else if(text.match(/>[ |\u00a0]/)) {
                handleRef();
            }
        }

    });
    // 将光标设置在某个元素后面
    function setCursorAfterElement(targetElement) {
        const range = document.createRange();
        const sel = window.getSelection();
        range.setStartAfter(targetElement);
        range.collapse(true);
        sel.removeAllRanges();
        sel.addRange(range);
    }
    editor.addEventListener('keydown', (e) => {
        if (e.ctrlKey && e.key === 's') {
            if(options.onsave) {
                e.preventDefault();
                options.onsave();
            }
        } else if (e.keyCode == 13) {
            // 回车
            var selection = window.getSelection();
            var anchorNode = selection.anchorNode;
            var text = anchorNode.textContent;
            if (text.startsWith("```")) {
                e.preventDefault();
                e.stopPropagation();
                originNode = anchorNode;
                if (originNode) {
                    var lang = "text";
                    if(originNode.textContent.startsWith("```")) {
                        lang = originNode.textContent.replaceAll("```", "");
                    } else {
                        var cl = originNode.children[0].classList;
                        for (let index = 0; index < cl.length; index++) {
                            const e = cl[index];
                            if(e.startsWith('language-')) {
                                lang = e.replace('language-', '');
                                break;
                            }
                        }
                    }
                    var pre = $c("pre");
                    var code = $c("code");
                    if(lang) {
                        code.setAttribute('class', `language-${lang}`);
                    }
                    var div = $c("div");
                    pre.appendChild(code);
                    div.appendChild(pre);
                    originNode.after(div);
                    var br = $c("br");
                    code.appendChild(br);
                    originNode.remove();
                    setCursorAfterElement(br);
                }
            } else if(e.ctrlKey) {
                // ctrl+回车 新行
                var div = $c("div");
                var br = $c("br");
                div.appendChild(br);
                editor.appendChild(div);
                setCursorAfterElement(br);
            }
        }
    });
    $$("loweditor-ap-strong").addEventListener("click", () => {
        applyStyle('strong');
    });
    $$("loweditor-ap-em").addEventListener("click", () => {
        applyStyle('em');
    });
    $$("loweditor-ap-table").addEventListener("click", () => {
        handleTable();
    });
    $$("loweditor-insert-newline").addEventListener("click", () => {
        editor.appendChild($c("br"));
    });
    return {
        editor,
        getHtml: function () {
            return this.editor.innerHTML;
        },
        setHtml: function (html) {
            this.editor.innerHTML = html;
        },
        setEditable,
        scrollTop: function(pos) {
            this.editor.scrollTop=pos;
        }
    };
}