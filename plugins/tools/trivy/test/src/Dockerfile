FROM node:12.0.0

# Using an old version of npm that has known vulnerabilities
RUN npm install -g npm@6.14.4

# Copy application files
WORKDIR /app
COPY package*.json ./

# Install dependencies with known vulnerabilities
RUN npm install express@4.16.4

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Start the application
CMD ["node", "app.js"] 