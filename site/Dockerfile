# Mímir site / mockup server
# Serves the landing page + app mockup. Run on a 5-digit port (never 8080):
#   docker build -t mimir-site site/
#   docker run -d -p 47312:80 --name mimir-site mimir-site
# Then open http://localhost:47312/app-mockup.html
FROM nginx:alpine
COPY index.html app-mockup.html /usr/share/nginx/html/
EXPOSE 80
