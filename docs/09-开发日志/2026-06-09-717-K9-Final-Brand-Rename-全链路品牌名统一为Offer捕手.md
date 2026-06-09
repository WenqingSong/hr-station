# K9-Final-Brand-Rename：全链路品牌名统一为 Offer捕手

> 日期：2026-06-09 | 性质：品牌文案统一替换

---

## 一、替换规则

| 旧文案 | 新文案 |
| ---- | ---- |
| Offerr鹅 | Offer捕手 |
| Offerr 鹅 | Offer捕手 |
| Offer鹅 | Offer捕手 |
| Offer 鹅 | Offer捕手 |

---

## 二、修改范围

| 目录 | 命中文件 | 替换方式 |
| ---- | ---- | ---- |
| web/src/ | 15 个 .tsx 文件 | sed 全局替换 |
| web/index.html | 1 文件 | sed |
| server/ | .go/.md/.sql 文件 | sed |
| docs/ | 全部 .md 文件 | sed |
| README.md | 1 文件 | sed |

---

## 三、替换结果

| 指标 | 结果 |
| ---- | ---- |
| web/src 新品牌 | 22 ✅ |
| docs 新品牌 | 200 ✅ |
| server 新品牌 | 25 ✅ |
| README 新品牌 | 2 ✅ |
| 旧品牌残留 | **0** ✅ |

---

## 四、构建

| 验收项 | 结果 |
| ---- | ---- |
| npm build | ✅ 2.03s |
| go build | ✅ |
| go vet | ✅ |

---

## 五、未修改范围确认

- ✅ 未修改数据库结构/API/业务逻辑/Navbar状态机/Modal/P2
