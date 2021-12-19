---
title = "Index"
---

Sessions:

<ul>
    {{ range $k, $page := .AllPages }}
    <li>
       <span class="post-date">{{ $page.CreatedAt }}</span>
       <a href="{{ $page.Url }}">{{ $page.Config.title }} - {{ $page.Config.title | reverse }}</a>
    </li>
    {{ end }}
</ul>
