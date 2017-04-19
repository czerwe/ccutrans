FROM scratch

ADD ccutrans /ccutrans
ADD mapping.json.example mapping.json

ENTRYPOINT ["/ccutrans"]