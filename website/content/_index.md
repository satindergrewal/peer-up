---
title: peer-up
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Open source &middot; Self-sovereign</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Access your home server&nbsp;<br class="hx:sm:block hx:hidden" />from anywhere
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  Peer-to-peer connectivity that works through NAT, CGNAT, and firewalls.&nbsp;<br class="hx:sm:block hx:hidden" />No accounts. No cloud. No trust required.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Get Started" link="docs/quick-start" >}}
</div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Works Through Any NAT"
    subtitle="Circuit relay v2 with private DHT. Verified on 5G CGNAT, double NAT, and corporate firewalls. If your device has internet, peer-up connects it."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(59,130,246,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Single Binary, Zero Dependencies"
    subtitle="One Go binary. No Docker, no Node.js, no database. Install it, run it, done. Works offline after the initial pairing."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(16,185,129,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="SSH-Style Trust Model"
    subtitle="An authorized_keys file decides who connects. You control the list. No accounts, no tokens, no central authority. Your network, your rules."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(245,158,11,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="60-Second Pairing"
    subtitle="Run peerup init on your server, peerup join on your laptop. One invite code, mutual authorization, done. Two commands from zero to connected."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(139,92,246,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Proxy Anything"
    subtitle="SSH, remote desktop, HTTP, databases — any TCP service tunneled through encrypted peer-to-peer streams. Access your home lab like you're on the same LAN."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(236,72,153,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Self-Healing Network"
    subtitle="Auto-reconnection with exponential backoff, config rollback if changes break connectivity, watchdog health monitoring. Your network fixes itself."
    style="background: radial-gradient(ellipse at 50% 80%,rgba(20,184,166,0.15),hsla(0,0%,100%,0));"
  >}}
{{< /hextra/feature-grid >}}
