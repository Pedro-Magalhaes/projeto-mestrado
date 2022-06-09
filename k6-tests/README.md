## Instalando k6 com integração com o kafka

Fonte: 

precisa do go 17+ e Git. Em seguida serguir os passos:

* go install go.k6.io/xk6/cmd/xk6@latest
Verifique que os binários do golang estão no path. No meu caso eles ficam em ~/go/bin/
* xk6 build --with github.com/mostafa/xk6-kafka@latestll
