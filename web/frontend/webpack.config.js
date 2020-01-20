var HtmlWebpackPlugin = require('html-webpack-plugin');
require("babel-polyfill");
const path = require('path');

module.exports = {
    mode: 'development',
    entry: ["babel-polyfill", path.resolve(__dirname, "src/index.js")],
    resolve: {
        extensions: ['.js', '.vue', '.png']
    },
    module: {
        rules: [
            {
                test: /\.vue?$/,
                exclude: /(node_modules)/,
                use: 'vue-loader'
            },
            {
                test: /\.js?$/,
                exclude: /(node_modules)/,
                use: 'babel-loader'
            },
            {
                test: /\.png?$/,
                exclude: /(node_modules)/,
                use: {
                    loader: 'url-loader',
                    options: {
                        limit: 10240000000,
                    },
                },
            }
        ]
    },
    plugins: [new HtmlWebpackPlugin({
        template: './src/index.html'
    })],
    devServer: {
        historyApiFallback: true
    },
    externals: {
        // global app config object
        config: JSON.stringify({
            apiUrl: 'http://localhost:4000'
        })
    }
}