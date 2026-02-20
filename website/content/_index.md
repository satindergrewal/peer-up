---
title: peer-up
layout: hextra-home
---

<div class="peerup-hero-logo">
  <img src="/peer-up/images/supermesh_logo_large.png" alt="peer-up logo" width="96" height="96" />
</div>

{{< hextra/hero-badge >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Open source &middot; Self-sovereign</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="peerup-hero-hook hx:mt-4 hx:mb-2">
  Your ISP put you behind CGNAT. Your VPN wants an account. Your cloud tunnel wants a subscription.
</div>

<div class="hx:mt-2 hx:mb-6">
{{< hextra/hero-headline >}}
  peer-up just&nbsp;<br class="hx:sm:block hx:hidden" />connects.
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-8">
{{< hextra/hero-subtitle >}}
  Peer-to-peer connectivity that works through NAT, CGNAT, and firewalls.&nbsp;<br class="hx:sm:block hx:hidden" />No accounts. No cloud. No trust required.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-12 hx:flex hx:flex-wrap hx:gap-3">
{{< hextra/hero-button text="Get Started" link="docs/quick-start" >}}
<a href="https://github.com/satindergrewal/peer-up" target="_blank" rel="noopener" class="peerup-secondary-btn">
  {{< icon name="github" attributes="height=18" >}}
  <span>View on GitHub</span>
</a>
</div>

<!-- ============================================================ -->
<!-- SECTION: Terminal Demo                                        -->
<!-- ============================================================ -->

<div class="peerup-section">
  <h2 class="peerup-section-title">From zero to connected in 60 seconds</h2>
  <p class="peerup-section-subtitle">Two commands on each device. No accounts to create, no keys to exchange manually, no ports to forward.</p>
  <div class="peerup-demo-container">
    <img src="/peer-up/images/terminal-demo.svg" alt="peer-up terminal demo showing init, invite, join, and proxy commands" class="peerup-demo-image" loading="lazy" />
  </div>
</div>

<!-- ============================================================ -->
<!-- SECTION: How It Works                                         -->
<!-- ============================================================ -->

<div class="peerup-section">
  <h2 class="peerup-section-title">How it works</h2>
  <p class="peerup-section-subtitle">Three steps. Both devices end up in each other's authorized_keys. That's it.</p>

  <div class="peerup-steps-grid">
    <div class="peerup-step">
      <div class="peerup-step-number">1</div>
      <img src="/peer-up/images/how-it-works-1-init.svg" alt="Step 1: Initialize peer-up on your server" class="peerup-step-image" loading="lazy" />
      <h3 class="peerup-step-title">Initialize</h3>
      <p class="peerup-step-desc">Run <code>peerup init</code> on your server. Generates a cryptographic identity, connects to the relay, and joins the private DHT.</p>
    </div>
    <div class="peerup-step">
      <div class="peerup-step-number">2</div>
      <img src="/peer-up/images/how-it-works-2-invite.svg" alt="Step 2: Create and share invite code" class="peerup-step-image" loading="lazy" />
      <h3 class="peerup-step-title">Invite</h3>
      <p class="peerup-step-desc">Run <code>peerup invite</code> to generate a one-time code. Share it through any channel — text, email, Signal, carrier pigeon.</p>
    </div>
    <div class="peerup-step">
      <div class="peerup-step-number">3</div>
      <img src="/peer-up/images/how-it-works-3-connect.svg" alt="Step 3: Join and start proxying services" class="peerup-step-image" loading="lazy" />
      <h3 class="peerup-step-title">Connect</h3>
      <p class="peerup-step-desc">Run <code>peerup join</code> on your laptop. Mutual authorization happens automatically. Proxy any TCP service through the encrypted tunnel.</p>
    </div>
  </div>
</div>

<!-- ============================================================ -->
<!-- SECTION: Feature Grid (existing, unchanged)                   -->
<!-- ============================================================ -->

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

<!-- ============================================================ -->
<!-- SECTION: Network Diagram                                      -->
<!-- ============================================================ -->

<div class="peerup-section">
  <h2 class="peerup-section-title">Direct when possible, relayed when necessary</h2>
  <p class="peerup-section-subtitle">peer-up uses encrypted circuit relay v2 to punch through NAT and CGNAT. When a direct path exists, it takes it. When it doesn't, the relay carries only encrypted streams — it never sees your data.</p>
  <div class="peerup-diagram-container">
    <img src="/peer-up/images/network-diagram.svg" alt="Network diagram showing peer-to-peer connections through NAT with relay fallback" class="peerup-diagram-image" loading="lazy" />
  </div>
</div>

<!-- ============================================================ -->
<!-- SECTION: Install                                              -->
<!-- ============================================================ -->

<div class="peerup-section">
  <h2 class="peerup-section-title">Install</h2>
  <p class="peerup-section-subtitle">Single binary. No runtime dependencies. Build from source with Go.</p>

  <div class="peerup-install-container">

{{< tabs items="macOS,Linux,From Source" >}}
{{< tab >}}
```bash
# Clone and build
git clone https://github.com/satindergrewal/peer-up.git
cd peer-up
go build -ldflags="-s -w" -trimpath -o peerup ./cmd/peerup

# Move to PATH
sudo mv peerup /usr/local/bin/

# Verify
peerup version
```
{{< /tab >}}
{{< tab >}}
```bash
# Clone and build
git clone https://github.com/satindergrewal/peer-up.git
cd peer-up
go build -ldflags="-s -w" -trimpath -o peerup ./cmd/peerup

# Move to PATH
sudo mv peerup /usr/local/bin/

# Verify
peerup version
```
{{< /tab >}}
{{< tab >}}
```bash
# Requires Go 1.22+
git clone https://github.com/satindergrewal/peer-up.git
cd peer-up
go build -ldflags="-s -w" -trimpath -o peerup ./cmd/peerup

# Run directly
./peerup version
```
{{< /tab >}}
{{< /tabs >}}

  </div>
</div>

<!-- ============================================================ -->
<!-- SECTION: Bottom CTA                                           -->
<!-- ============================================================ -->

<div class="peerup-section peerup-cta-section">
  <div class="peerup-cta-grid">
    <a href="https://github.com/satindergrewal/peer-up" target="_blank" rel="noopener" class="peerup-cta-card">
      <div class="peerup-cta-icon">{{< icon name="github" attributes="height=28" >}}</div>
      <h3 class="peerup-cta-title">Star on GitHub</h3>
      <p class="peerup-cta-desc">Browse the source, open issues, contribute</p>
    </a>
    <a href="docs/quick-start" class="peerup-cta-card">
      <div class="peerup-cta-icon">{{< icon name="book-open" attributes="height=28" >}}</div>
      <h3 class="peerup-cta-title">Documentation</h3>
      <p class="peerup-cta-desc">Quick start, architecture, daemon API</p>
    </a>
    <a href="blog" class="peerup-cta-card">
      <div class="peerup-cta-icon">{{< icon name="pencil" attributes="height=28" >}}</div>
      <h3 class="peerup-cta-title">Blog</h3>
      <p class="peerup-cta-desc">Engineering updates and release notes</p>
    </a>
  </div>
</div>
