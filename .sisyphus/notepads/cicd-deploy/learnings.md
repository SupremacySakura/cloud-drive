# CI/CD Deploy Task Learnings

## 2026-05-02

### What Worked Well

1. **appleboy/ssh-action** is a solid choice for GitHub Actions SSH deployment
   - Supports `script_stop: true` to fail fast on errors
   - `env` section allows passing variables to remote script
   - Well-documented and widely used

2. **Concurrency control** prevents deployment race conditions
   - `concurrency: deploy-production` ensures only one deployment runs at a time
   - Critical for avoiding conflicts when multiple pushes happen quickly

3. **Health checks after deployment** catch issues early
   - Backend `/healthz` endpoint verification
   - Frontend root path check
   - Container logs output on failure for debugging

### Patterns to Remember

1. **Deploy job structure**:
   ```yaml
   deploy:
     needs: [docker]  # Depend on validation
     if: github.ref == 'refs/heads/main' && github.event_name == 'push'
     concurrency: deploy-production
     timeout-minutes: 15
   ```

2. **SSH action usage**:
   ```yaml
   - uses: appleboy/ssh-action@v1
     with:
       host: ${{ secrets.SERVER_HOST }}
       username: ${{ secrets.SERVER_USER }}
       key: ${{ secrets.SSH_PRIVATE_KEY }}
       script_stop: true
       script: |
         # deployment commands
     env:
       DEPLOY_DIR: ${{ secrets.DEPLOY_DIR }}
   ```

3. **7-step deployment script**:
   1. Check .env exists
   2. git pull
   3. docker compose up
   4. Wait for startup
   5. Backend health check
   6. Frontend health check
   7. Image cleanup

### Gotchas

1. **script vs command**: appleboy/ssh-action uses `script:` not `command:`
2. **Environment variables**: Must use `env:` block, not inline in script
3. **known_hosts**: Dynamic ssh-keyscan has security implications; for production, consider pre-storing fingerprints
4. **Evidence files**: Subagents may not create evidence files automatically; need to generate them post-completion

### Security Considerations

- Secrets are properly masked in GitHub Actions logs
- SSH key should have limited permissions on server
- .env file check prevents deploying with default credentials
- script_stop ensures failed commands don't continue silently

### Tools Verified

- Python yaml module for YAML validation
- grep for structure verification
- git diff for change tracking

## References

- Plan: `.sisyphus/plans/cicd-deploy.md`
- Workflow: `.github/workflows/ci.yml`
- Documentation: `README.md` (服务器部署章节)
