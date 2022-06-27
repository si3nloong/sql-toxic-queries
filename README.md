# Toxic Queries

<p>Machine: Macbook Pro</p>
<p>CPU: M1 Max</p>
<p>Memory: 32GB</p>
<p>OS: macOS Monterey</p>
<p>Record set: 200000</p>

## Benchmarks

| Toxic Query             | Operation       |
| ----------------------- | --------------- |
| Count                   | 0.006944 ns/op  |
| Count with Explain      | 0.0001292 ns/op |
| With Leading %          | 0.001870 ns/op  |
| Without Leading %       | 0.001372 ns/op  |
| Offset Based Pagination | 0.04177 ns/op   |
| Cursor Based Pagination | 0.01518 ns/op   |
