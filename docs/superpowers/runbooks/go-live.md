# Go-live runbook — cli.craftybase.dev

These are manual, one-time steps (require repo-admin + Bunny DNS access).

1. **Merge** `feat/public-site` into `main`.
2. **Make the repo public:** GitHub → Settings → General → Danger Zone → Change visibility → Public. (Free GitHub Pages requires this.)
3. **Enable Pages via Actions:** Settings → Pages → Build and deployment → Source = **GitHub Actions**.
4. **Trigger the deploy:** push to `main` (or run the "Deploy site" workflow via *Run workflow*). Confirm it succeeds and publishes `website/dist`.
5. **Add the custom domain:** Settings → Pages → Custom domain → `cli.craftybase.dev` → Save. GitHub writes/uses the `public/CNAME` value and begins HTTPS provisioning.
6. **DNS on Bunny.net:** add a record `cli` (host) → **CNAME** → `craftybase.github.io`. Wait for propagation; GitHub will show "DNS check successful" and issue the certificate.
7. **Verify end-to-end:**
   - `https://cli.craftybase.dev/` and `/getting-started/` load over HTTPS.
   - `https://cli.craftybase.dev/reference/craftybase/` and `/llms.txt` resolve.
   - `curl -fsSL https://cli.craftybase.dev/install | bash` installs the binary from a real release (requires at least one published GoReleaser release; cut one if needed).
8. **(Optional follow-up)** repoint the CLI's "Learn more" footer: change `brand.DocsURL` to `https://cli.craftybase.dev/getting-started` (currently `https://craftybase.com/docs/api`).
