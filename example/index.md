---
title = "Index"
---

Sessions:

<ul>
    {{ range $k, $page := .AllPages }}
    <li>
       <span class="post-date">{{ $page.CreatedAt }}</span>
       <a href="{{ $page.Url }}">{{ $page.Config.title }}</a>
    </li>
    {{ end }}
</ul>
