FROM library/golang

# Godep for vendoring
# RUN go get github.com/tools/godep

# Recompile the standard library without CGO
RUN CGO_ENABLED=0 go install -a std

ENV APP_DIR $GOPATH/src/pigeon
RUN mkdir -p $APP_DIR


ADD . $APP_DIR

RUN go get github.com/spf13/viper && go get github.com/fsnotify/fsnotify 

# Compile the binary and statically link
RUN cd $APP_DIR && go build  

EXPOSE 9501
# Set the entrypoint
ENTRYPOINT (cd $APP_DIR && ./pigeon)